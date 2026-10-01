// Catalina-Support: backend de una mesa de ayuda de dos niveles.
//
// El arranque hace, en este orden: leer configuración, comprobar que se puede escribir
// el log, conectar con la base de datos, montar las rutas y escuchar. Cualquier fallo
// de esos pasos detiene el arranque: es preferible no arrancar a arrancar a medias.
//
// Las rutas de los módulos se registran con el prefijo /api/<módulo>/... y nginx las
// pasa tal cual (no las reescribe), así que el backend es el dueño de ese prefijo.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/juanky-estevez/go-logs"
	"gorm.io/gorm"

	airepositories "catalina-support/backend/modules/ai/repositories"
	aiservices "catalina-support/backend/modules/ai/services"
	authcontrollers "catalina-support/backend/modules/auth/controllers"
	authrepositories "catalina-support/backend/modules/auth/repositories"
	authservices "catalina-support/backend/modules/auth/services"
	"catalina-support/backend/modules/mail/controllers"
	"catalina-support/backend/modules/mail/repositories"
	"catalina-support/backend/modules/mail/services"
	settingscontrollers "catalina-support/backend/modules/settings/controllers"
	settingsrepositories "catalina-support/backend/modules/settings/repositories"
	settingsservices "catalina-support/backend/modules/settings/services"
	ticketcontrollers "catalina-support/backend/modules/tickets/controllers"
	ticketrepositories "catalina-support/backend/modules/tickets/repositories"
	ticketservices "catalina-support/backend/modules/tickets/services"
	usercontrollers "catalina-support/backend/modules/users/controllers"
	userrepositories "catalina-support/backend/modules/users/repositories"
	userservices "catalina-support/backend/modules/users/services"
	"catalina-support/backend/shared/auth"
	"catalina-support/backend/shared/authz"
	"catalina-support/backend/shared/config"
	"catalina-support/backend/shared/database"
	"catalina-support/backend/shared/httpx"
	"catalina-support/backend/shared/middleware"
	// **La base de zonas horarias va incrustada en el binario**: así `America/Guayaquil` se puede
	// cargar en cualquier contenedor, aunque no traiga `/usr/share/zoneinfo` (decisión 15).
	_ "time/tzdata"
)

const shutdownTimeout = 15 * time.Second

// ajustesDeTickets traduce la configuración de la instalación a lo que `tickets` necesita.
//
// El prefijo y el reparto viven en `settings` —son de la instalación, no de los tickets— y `tickets`
// los lee en cada ticket nuevo. El adaptador vive aquí, en el único sitio donde los dos módulos se
// conocen, y así ninguno de los dos importa al otro (docs/arquitectura.md, sección 4).
type ajustesDeTickets struct {
	settings *settingsservices.Service
}

// TicketSettings devuelve el prefijo y el reparto, tal y como los ve el módulo de tickets.
func (a ajustesDeTickets) TicketSettings() (ticketservices.TicketSettings, error) {
	config, err := a.settings.Config()
	if err != nil {
		return ticketservices.TicketSettings{}, err
	}

	return ticketservices.TicketSettings{
		NumberPrefix:         config.NumberPrefix,
		MainAssignment:       config.MainAssignment,
		MainNotification:     config.MainNotification,
		InternalAssignment:   config.InternalAssignment,
		InternalNotification: config.InternalNotification,
	}, nil
}

// proberDeCorreo traduce la prueba del correo saliente que pide el módulo de configuración a lo que
// ofrece el módulo de correo. Vive aquí, en el único sitio donde los dos módulos se conocen, y así
// **ninguno de los dos importa al otro** (docs/arquitectura.md, sección 4).
type proberDeCorreo struct {
	mail *services.Service
}

// ProbarCorreo prueba la conexión y la autenticación **sin mandar ningún correo**.
func (p proberDeCorreo) ProbarCorreo(correo settingsservices.CorreoSaliente) error {
	return p.mail.Probar(services.CorreoSaliente{
		Host:      correo.Host,
		Port:      correo.Port,
		Secure:    correo.Secure,
		User:      correo.User,
		Password:  correo.Password,
		FromName:  correo.FromName,
		FromEmail: correo.FromEmail,
	})
}

// resumidorDeIA traduce lo que `tickets` pide a lo que el módulo `ai` ofrece.
//
// Vive aquí, en el único sitio donde los dos módulos se conocen, y así **ninguno de los dos importa al
// otro** —la misma forma que `ajustesDeTickets`— (docs/arquitectura.md, sección 4 y
// docs/modules/ai.md, sección 5).
type resumidorDeIA struct {
	ai *aiservices.Service
}

// Pedir encola un resumen. **No espera al motor**: quien crea un ticket no puede quedarse mirando un
// contenedor (docs/modules/ai.md, decisión 2).
func (r resumidorDeIA) Pedir(entrada ticketservices.EntradaDeResumen) {
	r.ai.Pedir(aiservices.Entrada{
		Numero: entrada.Numero,
		Tipo:   aiservices.Tipo(entrada.Tipo),
		Texto:  entrada.Texto,
	})
}

// De devuelve los dos campos de esos tickets, en bloque.
func (r resumidorDeIA) De(numeros []string) (map[string]ticketservices.Resumenes, error) {
	resumenes, err := r.ai.De(numeros)
	if err != nil {
		return nil, err
	}

	convertidos := make(map[string]ticketservices.Resumenes, len(resumenes))
	for numero, dos := range resumenes {
		convertidos[numero] = ticketservices.Resumenes{
			Motivo:       resumenDelMotor(dos.Motivo),
			UltimaAccion: resumenDelMotor(dos.UltimaAccion),
		}
	}

	return convertidos, nil
}

// textosDeTickets le da al módulo `ai` el texto de un ticket, que lo arma `tickets`.
//
// Es el camino de vuelta del adaptador: `ai` declara que necesita un texto y aquí se le da, sin que
// los dos módulos se conozcan (docs/arquitectura.md, sección 4).
type textosDeTickets struct {
	tickets *ticketservices.Service
}

// TextoParaElMotor arma otra vez el texto de ese ticket.
func (t textosDeTickets) TextoParaElMotor(numero string) (string, error) {
	return t.tickets.TextoParaElMotor(numero)
}

// resumenDelMotor traduce un resumen del módulo `ai` al que espera `tickets`.
func resumenDelMotor(resumen aiservices.Resumen) ticketservices.Resumen {
	return ticketservices.Resumen{
		Estado:   string(resumen.Estado),
		Es:       resumen.Es,
		En:       resumen.En,
		ErrorKey: resumen.ErrorKey,
	}
}

func main() {
	if err := run(); err != nil {
		logs.LogError(err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// go-logs crea la carpeta por su cuenta y, si no puede escribir, NO falla: avisa en
	// rojo y sigue escribiendo sólo en terminal. Un LOGS_FOLDER mal puesto perdería el
	// log sin detener nada, así que aquí se comprueba de verdad.
	if err := os.MkdirAll(cfg.LogsFolder, 0o755); err != nil {
		return fmt.Errorf("no se puede escribir en LOGS_FOLDER (%s): %w", cfg.LogsFolder, err)
	}
	if probe, err := os.CreateTemp(cfg.LogsFolder, ".probe-*"); err != nil {
		return fmt.Errorf("no se puede escribir en LOGS_FOLDER (%s): %w", cfg.LogsFolder, err)
	} else {
		probe.Close()
		os.Remove(probe.Name())
	}

	logs.LogInfo("arrancando el backend en el entorno " + cfg.Environment)

	db, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("no se pudo conectar a la base de datos: %w", err)
	}
	logs.LogSuccess("conexión a la base de datos establecida")

	// El token de sesión: sin secreto no se firma nada, y en producción el arranque no sigue.
	tokens, err := auth.NewTokenManager(cfg.TokenSecret)
	if err != nil {
		return fmt.Errorf("no se puede firmar la sesión: %w", err)
	}

	// Los servicios, antes que las rutas: los middlewares que protegen cada ruta necesitan saber leer
	// quién llama, y eso vive en el módulo de cuentas.
	//
	// Los módulos `users` y `auth` se necesitan el uno al otro —`users` pide a `auth` el enlace del
	// alta, y `auth` pide a `users` la cuenta—, así que el círculo se rompe **aquí**, en un solo
	// sitio y a la vista: se construye el de cuentas y se le conecta después el de contraseñas. Se
	// rompe con una interfaz declarada por quien la usa, de modo que ninguno de los dos módulos
	// importa al otro (docs/arquitectura.md, sección 4).
	mailService := services.NewService(repositories.NewTemplateRepository(db), services.NewSender(), cfg.PublicAppURL)

	settingsService := settingsservices.NewService(settingsrepositories.NewSettingsRepository(db), cfg.FilesPath)

	usersService := userservices.NewService(userrepositories.NewUserRepository(db))
	authService := authservices.NewService(
		usersService,
		authrepositories.NewPasswordTokenRepository(db),
		mailService,
		tokens,
		authservices.Config{
			AdminPassword: cfg.AdminPassword,
			PublicAppURL:  cfg.PublicAppURL,
			TokenSecret:   cfg.TokenSecret,
		},
	)
	usersService.SetPasswordLinks(authService)
	// **Los enlaces, la vuelta de Keycloak y la zona horaria salen de la configuración** (decisiones 14
	// y 15), con la variable de entorno como respaldo: cambiarlos en Configuración vale sin reiniciar nada.
	mailService.SetInstalacion(instalacionesDeLaConfiguracion{settings: settingsService})
	authService.SetEnlaces(instalacionesDeLaConfiguracion{settings: settingsService})

	// **Cómo se entra en esta instalación sale de la base, no del entorno**: el método, el directorio y
	// el reino se guardan desde Configuración y se leen en cada intento, así que cambiarlos vale sin
	// reiniciar nada (docs/modules/settings.md, sección 5.8). Las dos interfaces se declaran en quien las
	// usa y se conectan aquí, que es el único sitio donde los módulos se conocen.
	authService.SetAccess(settingsService)

	// Las dos pruebas de conexión de Configuración las contesta `auth`, que es quien sabe hablar con un
	// directorio y con un reino: probar una conexión es hablar el protocolo, y eso no se duplica.
	settingsService.SetProber(authService)

	// Y la prueba del correo saliente la contesta `mail`, con la misma conexión y autenticación que
	// usa para enviar: **sin mandar ningún correo**, que es lo que hace falta en el asistente
	// (docs/primer-arranque.md, sección 3).
	settingsService.SetProberDeCorreo(proberDeCorreo{mail: mailService})

	// Reactivar una cuenta de AD pregunta al directorio si esa persona sigue allí, y la pregunta la
	// contesta `auth` **con la configuración guardada**: la misma búsqueda y la misma cuenta de servicio
	// con las que se entra (docs/modules/users.md, sección 5, punto 4).
	usersService.SetDirectory(authService)

	// El idioma de la instalación es el que se le pone a una cuenta cuando nadie le elige uno.
	usersService.SetInstallation(settingsService)

	// Los tickets: el turno y los nombres salen de `users`, la configuración de `settings` y los avisos
	// de `mail`. Los tres se conectan aquí, que es el único sitio donde los módulos se conocen
	// (docs/arquitectura.md, sección 4).
	ticketsService := ticketservices.NewService(
		ticketrepositories.NewTicketRepository(db),
		ticketrepositories.NewConversationRepository(db),
		ticketrepositories.NewCategoryRepository(db),
		cfg.FilesPath,
	)
	ticketsService.SetAccounts(usersService)
	ticketsService.SetMailer(mailService)
	ticketsService.SetConfiguration(ajustesDeTickets{settings: settingsService})

	// El motor de IA: redacta el motivo y la última acción de cada ticket (docs/modules/ai.md). **Es
	// opcional a propósito**: sin `AI_URL` no se engancha nada y la mesa de ayuda funciona entera, con
	// los dos campos sin texto. Un contenedor caído no puede parar el producto.
	aiService := aiservices.NewService(
		airepositories.NewAIRepository(db),
		aiservices.Ajustes{
			URL:      cfg.AIURL,
			Modelo:   cfg.AIModel,
			Contexto: 4096,
			Palabras: cfg.AIPalabras,
			Espera:   cfg.AIEspera,
		},
	)
	ticketsService.SetInsights(resumidorDeIA{ai: aiService})

	// **El motor sale de la configuración, no del entorno**: su dirección y su modelo se guardan
	// desde Configuración y se leen en cada petición, así que cambiarlos vale sin reiniciar nada. Lo
	// del entorno queda como respaldo para una instalación que ya lo tuviera puesto así
	// (docs/modules/ai.md).
	aiService.SetConfiguracion(instalacionesDeLaConfiguracion{settings: settingsService})

	// El motor también necesita poder **volver a armar el texto** de un ticket: es lo que hace la
	// puesta al día del arranque, cuando la cola se ha perdido y hay resúmenes a medias
	// (docs/modules/ai.md, decisión 10).
	aiService.SetFuente(textosDeTickets{tickets: ticketsService})

	// Al arrancar se vuelve a pedir lo que quedó a medias: la cola vive en memoria y se perdió al
	// parar (docs/modules/ai.md, decisión 10).
	aiService.Retomar()

	// Quién llama, en cada petición: el cargador lee la cuenta de la base.
	whoIsCalling := loadIdentity(usersService)

	// Los cuatro permisos que existen hoy: administrador, personal (Soporte o Administrador), y
	// cualquiera que haya entrado.
	admin := func(handler http.Handler) http.Handler {
		return middleware.Chain(handler,
			middleware.Auth(tokens, whoIsCalling),
			authz.RequireRole(auth.RoleAdministrador),
		)
	}
	staff := func(handler http.Handler) http.Handler {
		return middleware.Chain(handler,
			middleware.Auth(tokens, whoIsCalling),
			authz.RequireRole(auth.RoleAdministrador, auth.RoleSoporte),
		)
	}
	authenticated := func(handler http.Handler) http.Handler {
		return middleware.Chain(handler, middleware.Auth(tokens, whoIsCalling))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health(db))

	// El editor de plantillas de correo: los cuatro endpoints son de administrador
	// (docs/modules/mail.md, sección 7).
	mailController := controllers.NewTemplateController(mailService)

	mux.Handle("GET /api/mail/templates", admin(http.HandlerFunc(mailController.List)))
	mux.Handle("PUT /api/mail/templates/{key}/{language}", admin(http.HandlerFunc(mailController.Update)))
	mux.Handle("POST /api/mail/templates/{key}/{language}/reset", admin(http.HandlerFunc(mailController.Reset)))
	mux.Handle("POST /api/mail/templates/{key}/{language}/test", admin(http.HandlerFunc(mailController.Test)))
	// La vista previa la renderiza el backend, con los mismos datos de ejemplo que la prueba
	// (docs/modules/mail.md, decisión 17).
	mux.Handle("POST /api/mail/templates/{key}/{language}/preview", admin(http.HandlerFunc(mailController.Preview)))

	usersController := usercontrollers.NewUserController(usersService)
	authController := authcontrollers.NewAuthController(authService)
	// **El motor de IA es opcional**, y el asistente lo dice en su resumen: el cableado le da al módulo
	// de configuración con qué preguntarlo (docs/primer-arranque.md, sección 4).
	settingsService.SetProberDeIA(aiService)
	setupController := settingscontrollers.NewSetupController(settingsService)
	settingsController := settingscontrollers.NewSettingsController(settingsService)

	// Entrar y pedir el enlace son públicos: son justo los sitios por los que se pasa cuando todavía
	// no se ha entrado (docs/modules/auth.md, sección 8).
	// Los caminos de entrada, que la pantalla lee antes de que nadie haya entrado: es público y no
	// lleva ningún dato de nadie.
	mux.HandleFunc("GET /api/auth/methods", authController.Methods)
	mux.HandleFunc("POST /api/auth/login", authController.Login)
	mux.HandleFunc("POST /api/auth/password/forgot", authController.Forgot)
	mux.HandleFunc("POST /api/auth/password/reset", authController.Reset)

	// El camino de Keycloak: los dos son públicos a propósito. El primero porque lo pide el navegador
	// antes de entrar, y el segundo porque lo llama **Keycloak**, que no lleva nuestra cabecera de
	// sesión: lo que protege la vuelta es el `state` firmado, no un token
	// (docs/modules/auth.md, secciones 5.3 y 8).
	mux.HandleFunc("GET /api/auth/keycloak/start", authController.KeycloakStart)
	mux.HandleFunc("GET /api/auth/keycloak/callback", authController.KeycloakCallback)

	// **La vista de primer arranque**: no lleva sesión —todavía no hay puerta por la que entrar— y lo
	// que la protege es el sello: en cuanto la instalación está sellada, todas contestan 409
	// (docs/primer-arranque.md, sección 6).
	mux.HandleFunc("GET /api/setup", setupController.Show)
	mux.HandleFunc("POST /api/setup/installation", setupController.SaveInstallation)
	mux.HandleFunc("POST /api/setup/entry", setupController.SaveEntry)
	mux.HandleFunc("POST /api/setup/location", setupController.SaveLocation)
	mux.HandleFunc("POST /api/setup/mail", setupController.SaveMail)
	// Las dos pruebas de conexión del asistente: se prueba **lo que hay en pantalla**, antes de
	// guardarlo, y por eso van por `POST` con los datos en el cuerpo y no tocan la base. El candado es
	// el mismo que el del resto del asistente: sellada la instalación, contestan **409**
	// (docs/primer-arranque.md, sección 6).
	mux.HandleFunc("POST /api/setup/entry/test", setupController.TestEntry)
	mux.HandleFunc("POST /api/setup/mail/test", setupController.TestMail)
	mux.HandleFunc("POST /api/setup/finish", setupController.Finish)

	// El resto de la sesión exige haber entrado, sin mirar el papel.
	mux.Handle("GET /api/auth/me", authenticated(http.HandlerFunc(authController.Me)))
	mux.Handle("POST /api/auth/logout", authenticated(http.HandlerFunc(authController.Logout)))
	mux.Handle("POST /api/auth/password/change", authenticated(http.HandlerFunc(authController.Change)))

	// Dar de alta: Soporte sólo puede crear `usuario`, y eso se comprueba en el servicio, que es quien
	// sabe con qué datos se está intentando (docs/modules/users.md, sección 5).
	mux.Handle("POST /api/users", staff(http.HandlerFunc(usersController.Create)))
	mux.Handle("POST /api/users/{id}/reset-password", staff(http.HandlerFunc(usersController.ResetPassword)))

	// La lista la ven Soporte y Administrador; la ficha de una cuenta, sólo un Administrador, y los
	// cambios de papel y de origen también (docs/modules/users.md, sección 4).
	mux.Handle("GET /api/users", staff(http.HandlerFunc(usersController.List)))
	mux.Handle("GET /api/users/{id}", admin(http.HandlerFunc(usersController.Show)))
	mux.Handle("PATCH /api/users/{id}", staff(http.HandlerFunc(usersController.Update)))
	mux.Handle("POST /api/users/{id}/deactivate", staff(http.HandlerFunc(usersController.Deactivate)))
	mux.Handle("POST /api/users/{id}/activate", staff(http.HandlerFunc(usersController.Activate)))
	mux.Handle("POST /api/users/{id}/origin", admin(http.HandlerFunc(usersController.ChangeOrigin)))

	// El perfil propio lo cambia cualquiera que haya entrado, y sólo lo suyo.
	mux.Handle("PATCH /api/users/me", authenticated(http.HandlerFunc(usersController.UpdateProfile)))

	// Los tickets. Aquí la ruta sólo exige haber entrado: **quién puede qué depende del ticket** —un
	// usuario ve el suyo y no el de otro, y no ve ningún interno—, y eso lo decide el servicio, que es
	// quien ve el ticket y a quien lo pide (docs/usuarios-y-permisos.md, sección 3).
	ticketsController := ticketcontrollers.NewTicketController(ticketsService)

	mux.Handle("GET /api/tickets", authenticated(http.HandlerFunc(ticketsController.List)))
	mux.Handle("POST /api/tickets", authenticated(http.HandlerFunc(ticketsController.Create)))
	// La lista de responsables posibles: la da `tickets` para que la pantalla no tenga que llamar a
	// `/api/users` (docs/arquitectura.md, sección 4).
	mux.Handle("GET /api/tickets/assignees", authenticated(http.HandlerFunc(ticketsController.Assignees)))
	mux.Handle("GET /api/tickets/{number}", authenticated(http.HandlerFunc(ticketsController.Show)))
	mux.Handle("PATCH /api/tickets/{number}", authenticated(http.HandlerFunc(ticketsController.Update)))
	mux.Handle("POST /api/tickets/{number}/assign", authenticated(http.HandlerFunc(ticketsController.Assign)))
	mux.Handle("POST /api/tickets/{number}/state", authenticated(http.HandlerFunc(ticketsController.State)))
	mux.Handle("POST /api/tickets/{number}/escalate", authenticated(http.HandlerFunc(ticketsController.Escalate)))
	mux.Handle("POST /api/tickets/{number}/reopen", authenticated(http.HandlerFunc(ticketsController.Reopen)))
	// «Regenerar» los dos campos que redacta el motor de IA: es de `tickets` y no de `ai`, porque la
	// pantalla de tickets sólo habla con su propia API (docs/modules/ai.md, sección 5).
	mux.Handle("POST /api/tickets/{number}/insights", authenticated(http.HandlerFunc(ticketsController.Insights)))
	// El catálogo de categorías y las etiquetas (docs/modules/tickets.md, decisiones 64 a 68). Los
	// caminos literales `categories` y `tags` **no chocan con `{number}`**: el enrutador de Go 1.22
	// elige el patrón más específico, y un segmento literal gana a un comodín.
	mux.Handle("GET /api/tickets/categories", authenticated(http.HandlerFunc(ticketsController.ListCategories)))
	mux.Handle("POST /api/tickets/categories", authenticated(http.HandlerFunc(ticketsController.CreateCategory)))
	mux.Handle("PATCH /api/tickets/categories/{id}", authenticated(http.HandlerFunc(ticketsController.RenameCategory)))
	mux.Handle("POST /api/tickets/categories/{id}/state", authenticated(http.HandlerFunc(ticketsController.CategoryState)))
	mux.Handle("GET /api/tickets/tags", authenticated(http.HandlerFunc(ticketsController.Tags)))
	// El mantenimiento de las etiquetas: el catálogo se crea, se renombra —y el cambio vale para todos
	// los tickets que la llevan— y se retira —y se quita de ellos— (docs/modules/tickets.md, decisión
	// 72). El `{tag}` de la ruta es el nombre normalizado, que es único en el catálogo, y va en un
	// segmento: una etiqueta nunca lleva barra, así que no se come nada de la ruta.
	mux.Handle("POST /api/tickets/tags", authenticated(http.HandlerFunc(ticketsController.CreateTag)))
	mux.Handle("PATCH /api/tickets/tags/{tag}", authenticated(http.HandlerFunc(ticketsController.RenameTag)))
	mux.Handle("DELETE /api/tickets/tags/{tag}", authenticated(http.HandlerFunc(ticketsController.DeleteTag)))
	// Los observadores se añaden de dos maneras —una mención en un comentario o el botón de la ficha
	// desde el 2026-09-28— y las dos dejan lo mismo. Añadir y quitar son de Soporte y Desarrollo
	// (decisión 63), y el servicio comprueba las dos cosas: quién lo pide y que la cuenta valga.
	// **El `{id}` de la ruta de quitar es el identificador de la cuenta**, no el de la fila de la
	// tabla: quien se quita del ticket es una persona, y la fila es cosa de dentro
	// (`docs/modules/tickets.md`, decisiones 63 y 74).
	mux.Handle("POST /api/tickets/{number}/observers", authenticated(http.HandlerFunc(ticketsController.AddObserver)))
	mux.Handle("DELETE /api/tickets/{number}/observers/{id}", authenticated(http.HandlerFunc(ticketsController.RemoveObserver)))
	mux.Handle("POST /api/tickets/{number}/comments", authenticated(http.HandlerFunc(ticketsController.Comment)))
	mux.Handle("PATCH /api/tickets/{number}/comments/{id}", authenticated(http.HandlerFunc(ticketsController.EditComment)))
	mux.Handle("DELETE /api/tickets/{number}/comments/{id}", authenticated(http.HandlerFunc(ticketsController.DeleteComment)))
	mux.Handle("POST /api/tickets/{number}/attachments", authenticated(http.HandlerFunc(ticketsController.Upload)))
	mux.Handle("GET /api/tickets/{number}/attachments/{id}", authenticated(http.HandlerFunc(ticketsController.Download)))

	// La configuración de la instalación y su marca. Lo que ve un administrador va detrás de `admin`.
	mux.Handle("GET /api/settings", admin(http.HandlerFunc(settingsController.Show)))
	mux.Handle("PUT /api/settings", admin(http.HandlerFunc(settingsController.Update)))
	mux.Handle("POST /api/settings/brand/logo", admin(http.HandlerFunc(settingsController.UploadLogo)))
	// Las dos pruebas de conexión: se prueba **lo que hay en pantalla**, antes de guardarlo, y por eso
	// van por `POST` con la configuración en el cuerpo y no leen nada de la base
	// (docs/modules/settings.md, sección 5.8).
	mux.Handle("POST /api/settings/directory/test", admin(http.HandlerFunc(settingsController.TestDirectory)))
	mux.Handle("POST /api/settings/keycloak/test", admin(http.HandlerFunc(settingsController.TestKeycloak)))
	// La prueba del motor de IA: pregunta a su comprobación de salud la dirección que se le manda, o
	// la que hay guardada si no llega ninguna (docs/modules/ai.md).
	mux.Handle("POST /api/settings/ai/test", admin(http.HandlerFunc(settingsController.TestAI)))
	mux.Handle("DELETE /api/settings/brand/logo", admin(http.HandlerFunc(settingsController.DeleteLogo)))

	// Y la marca que necesita la aplicación **antes de que nadie entre**: el color institucional y el
	// logo. Son públicos a propósito, y sólo devuelven la marca de la institución
	// (docs/modules/settings.md, sección 5.4).
	mux.HandleFunc("GET /api/settings/brand", settingsController.Brand)
	mux.HandleFunc("GET /api/settings/brand/logo", settingsController.LogoFile)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           middleware.Chain(mux, middleware.Recover),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// El servidor escucha en una goroutine para que main pueda esperar la señal de parada.
	serverErrors := make(chan error, 1)
	go func() {
		logs.LogSuccess("escuchando en el puerto " + cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("el servidor se detuvo: %w", err)
	case signalReceived := <-stop:
		logs.LogWarning("señal recibida (" + signalReceived.String() + "): cerrando")
	}

	// Se deja terminar las peticiones en curso antes de salir.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("no se pudo cerrar el servidor con orden: %w", err)
	}

	// Se para la cola del motor antes de cerrar la base: un trabajo a medias escribiría en una conexión
	// cerrada.
	aiService.Parar()

	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("no se pudo cerrar la conexión a la base de datos: %w", err)
		}
	}

	logs.LogSuccess("backend detenido")
	return nil
}

// loadIdentity resuelve el `sub` de un token en la cuenta que hay detrás.
//
// Se lee la cuenta **en cada petición**, que es lo que hace que desactivar a alguien o cambiarle el
// papel surta efecto al instante, sin esperar a que caduque su token
// (docs/usuarios-y-permisos.md, sección 6).
//
// La cuenta de fábrica no se busca en la base: no está allí. Su token lleva el sujeto `admin` y su
// identidad sale de la configuración (docs/usuarios-y-permisos.md, sección 8).
func loadIdentity(users *userservices.Service) auth.IdentityLoader {
	return func(_ context.Context, subject string) (auth.Identity, error) {
		if subject == auth.FactorySubject {
			return auth.FactoryIdentity(), nil
		}

		id, err := strconv.ParseInt(subject, 10, 64)
		if err != nil {
			// Un sujeto que no es un identificador ni la cuenta de fábrica: el token no vale.
			return auth.Identity{}, auth.ErrIdentityNotFound
		}

		account, err := users.ByID(id)
		if errors.Is(err, auth.ErrAccountNotFound) {
			return auth.Identity{}, auth.ErrIdentityNotFound
		}
		if err != nil {
			return auth.Identity{}, err
		}

		// Desactivada es lo mismo que no existir: para quien la usa, su sesión ya no vale.
		if !account.IsActive {
			return auth.Identity{}, auth.ErrIdentityNotFound
		}

		return account.Identity(), nil
	}
}

// health comprueba que el proceso responde y que la base de datos contesta de verdad.
func health(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := db.DB()
		if err != nil {
			logs.LogError("health: no se pudo obtener la conexión a la base de datos: " + err.Error())
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "database": "error"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			logs.LogError("health: la base de datos no responde: " + err.Error())
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "database": "error"})
			return
		}

		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "ok"})
	}
}

// instalacionesDeLaConfiguracion es el cableado entre el módulo de configuración y los módulos que
// necesitan saber **cómo es esta instalación**: su dirección pública y su zona horaria. Cada módulo
// declara lo que necesita (`Instalacion`, `Enlaces`) y aquí se le da, sin que ninguno importe al otro
// (docs/arquitectura.md, sección 4).
type instalacionesDeLaConfiguracion struct {
	settings *settingsservices.Service
}

func (i instalacionesDeLaConfiguracion) DireccionPublica() string {
	return i.settings.DireccionPublica()
}

func (i instalacionesDeLaConfiguracion) ZonaHoraria() string { return i.settings.ZonaHoraria() }

// AI traduce el motor de la configuración al tipo que declara el módulo de IA, y **dice si hay uno
// puesto**: en falso, el módulo de IA usa el respaldo del entorno, que es como funcionaba antes
// (docs/modules/ai.md).
func (i instalacionesDeLaConfiguracion) AI() (aiservices.AIDatos, error) {
	motor, err := i.settings.AI()
	if err != nil {
		return aiservices.AIDatos{}, err
	}

	return aiservices.AIDatos{URL: motor.URL, Modelo: motor.Modelo, Hay: motor.Hay}, nil
}

// SMTP traduce el correo saliente de la configuración al tipo que declara el módulo de correo, y
// **dice si hay alguno puesto**: en falso, el remitente usa el del entorno.
func (i instalacionesDeLaConfiguracion) SMTP() (services.CorreoSaliente, bool) {
	correo, err := i.settings.SMTP()
	if err != nil || !correo.Set {
		return services.CorreoSaliente{}, false
	}

	return services.CorreoSaliente{
		Host:      correo.Host,
		Port:      correo.Port,
		Secure:    correo.Secure,
		User:      correo.User,
		Password:  correo.Password,
		FromName:  correo.FromName,
		FromEmail: correo.FromEmail,
	}, true
}
