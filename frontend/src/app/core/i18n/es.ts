/**
 * Los textos de la interfaz, en español.
 *
 * **Este archivo es la fuente del tipo**: `Textos` sale de aquí, así que el diccionario inglés
 * está obligado a tener las mismas claves. Una clave que falte en un idioma **no compila**, que
 * es la única forma de que la regla «cada clave necesita su texto en los dos idiomas» no dependa
 * de que alguien se acuerde (docs/modules/auth.md, decisión 26).
 *
 * Los textos de error van en el mismo diccionario: son textos de interfaz como los demás.
 */
export const ES = {
  idioma: {
    nombre: 'Español',
    cambiar: 'Cambiar idioma',
    es: 'Español',
    en: 'English',
  },

  configuracion: {
    titulo: 'Configuración',
    instalacion: 'La instalación',
    instalacionAyuda:
      'La institución, la empresa o el equipo, y el idioma con el que habla. El nombre es lo que se lee en la pantalla de entrada, en el menú y en la pestaña del navegador.',
    nombreCampo: 'Nombre',
    nombreDeFabrica: 'Si lo dejas vacío, vuelve el nombre de fábrica (Catalina Support).',
    guardarInstalacion: 'Guardar la instalación',
    instalacionGuardada: 'La instalación se ha guardado.',
    marca: 'La marca',
    marcaAyuda:
      'El logo y el color de la institución. El logo se enseña en la pantalla de entrada y en el menú, y el color tiñe los dos temas de fábrica.',
    logoClaro: 'Logo para los temas claros',
    logoClaroAyuda: 'Se enseña en Claro, Papel, Niebla y Alto contraste.',
    logoOscuro: 'Logo para los temas oscuros',
    logoOscuroAyuda:
      'Se enseña en Oscuro, Grafito, Noche y Sepia. Si no pones ninguno, se usa el otro.',
    vistaPrevia: 'Vista previa',
    logoPuesto: 'Tienes un logo propio puesto.',
    logoDeFabrica: 'Se está usando el logo de fábrica.',
    elegirArchivo: 'Elegir un archivo',
    formatos: 'PNG, JPEG, WebP o SVG, hasta 1 MB y 2000 píxeles de lado.',
    subir: 'Subir el logo',
    subiendo: 'Subiendo…',
    volverAlDeFabrica: 'Volver al de fábrica',
    logoSubido: 'Logo subido. Ya se está usando.',
    logoQuitado: 'Ese hueco ha vuelto al logo de fábrica.',
    colorElegido: 'Color',
    guardarMarca: 'Guardar la marca',
    guardando: 'Guardando…',
    descartar: 'Descartar',
    marcaGuardada: 'La marca se ha guardado.',
    ejemploBoton: 'Un botón de ejemplo',
    ejemploTexto: 'y un texto al lado, para ver cómo queda el conjunto.',
    colorResuelto: 'En los temas claros queda',
    colorResueltoOscuro: 'y en los oscuros',
    idioma: 'El idioma de la instalación',
    idiomaAyuda:
      'El idioma de toda la instalación: interfaz, fechas, correos y textos generados por IA.',
    traduccionTitulo: 'Revisar el cambio de idioma',
    traduccionAyuda: 'Revisa y corrige las once plantillas antes de aplicar el idioma a toda la instalación.',
    traduccionCosto: 'El proveedor externo puede cobrar 22 solicitudes: estimación conservadora de 12 000 tokens de entrada y 8 000 de salida. Puedes aceptarlo o editar las plantillas manualmente.',
    traducirConfirmando: 'Aceptar costo y traducir',
    editarManualmente: 'Editar manualmente',
    asuntoTraducido: 'Asunto',
    cuerpoTraducido: 'Cuerpo',
    destinoExistente: 'Versión personalizada existente',
    aplicarIdioma: 'Aplicar idioma',
    correos: 'Los correos',
    correosAyuda:
      'Los diez correos que manda la aplicación, en los dos idiomas, se editan en su propia pantalla: es del módulo de correo, y aquí sólo se llega a ella.',
    editarCorreos: 'Editar los correos',
    region: 'Región horaria',
    regionAyuda:
      'La zona en la que se leen todas las fechas, aquí y en los correos. Se guardan en UTC: cambiarla no mueve ningún ticket.',
    regionSeleccionada: 'Región horaria seleccionada',
    buscarZona: 'Buscar una ciudad o una zona',
    sinZonas: 'Ninguna zona coincide con lo que buscas.',
    horaDeLaZona: 'Ahora son las {hora} ({desfase}).',
    regionYDireccion: 'Región horaria y dirección pública',
    regionYDireccionAyuda:
      'La zona en la que se leen las fechas y la dirección por la que se entra a esta instalación.',
    guardarRegion: 'Guardar la región y la dirección',
    regionGuardada: 'La región y la dirección se han guardado.',
    direccion: 'Dirección pública',
    direccionAyuda:
      'La dirección por la que se entra a esta instalación: es la base de los enlaces que van en los correos y la vuelta de Keycloak. Sirve http o https, con puerto si hace falta, y también localhost.',
    direccionEjemplo: 'https://soporte.mi-institucion.org',
    direccionInsegura:
      'Esta instalación no está sirviendo por https: la contraseña y la sesión viajan sin cifrar por la red. Se puede usar así —para probarla en local, por ejemplo—, y para usarla en serio conviene ponerle un certificado.',
    motorDeIA: 'El motor de IA',
    motorDeIAAyuda:
      'El motor que redacta el «Motivo» y la «Última acción» de cada ticket. Su configuración es obligatoria; si deja de responder, la mesa de ayuda continúa y recupera los resúmenes cuando vuelva.',
    iaDireccion: 'Dirección del motor',
    iaDireccionAyuda:
      'Dónde escucha el motor. Sirve http o https, con puerto si hace falta. Si se deja vacío, la instalación no tiene motor y el entorno hace de respaldo.',
    iaDireccionEjemplo: 'http://catalina_support_ai:8080',
    iaModelo: 'Modelo',
    iaModeloAyuda:
      'El modelo con el que redacta, tal y como lo nombra el motor. Si se deja vacío, se usa el del entorno.',
    iaModeloEjemplo: 'qwen2.5-1.5b-instruct',
    probarIA: 'Probar la conexión',
    guardarIA: 'Guardar el motor de IA',
    iaLocal: 'En este servidor',
    iaRemota: 'En otro servidor',
    iaProveedor: 'Proveedor externo',
    iaCompatible: 'Compatible con OpenAI',
    iaAutenticacion: 'Autenticación',
    iaAuth_none: 'Sin autenticación',
    iaAuth_bearer: 'Token Bearer',
    iaAuth_header: 'Clave en cabecera',
    iaAuth_basic: 'Usuario y contraseña',
    iaCabecera: 'Nombre de la cabecera',
    iaCredencial: 'Credencial o token',
    iaPrivacidad: 'Confirmo que los textos del ticket saldrán de este servidor.',
    iaPrivacidadCosto:
      'Confirmo que los textos del ticket saldrán de este servidor y que esta prueba puede generar costos del proveedor.',
    iaCatalogo: 'Modelos locales',
    iaCargarCatalogo: 'Ver y actualizar el catálogo',
    iaAceptarLicencia: 'Acepto la licencia de esta versión.',
    iaDescargar: 'Descargar',
    iaActivar: 'Activar',
    iaEliminar: 'Eliminar',
    iaInstalado: 'Instalado',
    iaActivo: 'Activo',
    iaSaludable: 'Motor disponible',
    iaNoSaludable: 'Motor reiniciándose',
    iaDescargando: 'Descargando',
    iaDiscoDisponible: 'Disco disponible: {cantidad}',
    iaRecursos: 'Descarga {disco} · RAM recomendada {ram}',
    motorOk: 'El motor ha contestado: redacta los dos resúmenes del ticket.',
    motorGuardado: 'El motor de IA se ha guardado.',
    entrada: 'Método de autenticación',
    entradaAyuda:
      'La instalación entra por un método a la vez: los otros dos quedan apagados aunque estén configurados. La cuenta de fábrica entra siempre, y desde ella se puede volver a cambiar si se elige mal.',
    metodo: 'Método de entrada',
    metodoLocal: 'Cuentas de la aplicación',
    metodoLocalAyuda: 'Las cuentas viven aquí y se entra con correo y contraseña.',
    metodoAD: 'Directorio de la organización (AD)',
    metodoADAyuda:
      'Se entra con la cuenta de la organización, contra el directorio. Las contraseñas locales no valen, ni siquiera para Soporte y Desarrollo.',
    metodoKeycloak: 'Keycloak',
    metodoKeycloakAyuda:
      'Se entra desde la pantalla de Keycloak, con su cuenta. Aquí no hay formulario de correo y contraseña.',
    metodoSinConfigurar: 'Completa los datos aquí abajo antes de guardar el método.',
    guardarEntrada: 'Guardar cómo se entra',
    entradaGuardada: 'Guardado. Vale desde la próxima entrada, sin reiniciar nada.',
    directorio: 'El directorio de la organización',
    directorioAyuda:
      'Lo que hace falta para entrar por AD. La contraseña de la cuenta de servicio se guarda y no se vuelve a enseñar: si se deja vacía, se queda la que había.',
    dirServidor: 'Servidor',
    dirServidorAyuda: 'El nombre o la dirección del directorio. Ejemplo: ldap o ad.empresa.local.',
    dirPuerto: 'Puerto',
    dirPuertoAyuda: '389 para LDAP y 636 para LDAPS. Si se deja vacío, 389.',
    dirTls: 'Conexión cifrada (LDAPS)',
    dirTlsAyuda: 'En producción, sí: por ahí viajan credenciales.',
    dirCuenta: 'Cuenta de servicio',
    dirCuentaAyuda:
      'El nombre completo con el que se busca a la gente. Ejemplo: cn=consulta,ou=servicios,dc=empresa,dc=local.',
    dirContrasena: 'Contraseña de la cuenta de servicio',
    dirContrasenaPuesta: 'Hay una contraseña guardada. Déjalo vacío para no cambiarla.',
    dirContrasenaVacia: 'No hay ninguna contraseña guardada.',
    dirBase: 'Dónde se busca',
    dirBaseAyuda: 'La base de búsqueda. Ejemplo: ou=personas,dc=empresa,dc=local.',
    dirFiltro: 'Filtro de búsqueda',
    dirFiltroAyuda:
      'Con «%s» donde va el correo de quien entra. Ejemplo: (mail=%s) o (userPrincipalName=%s).',
    dirAtributoCorreo: 'Atributo del correo',
    dirAtributoNombre: 'Atributo del nombre',
    dirAtributoApellidos: 'Atributo de los apellidos',
    dirAtributoId: 'Atributo del identificador',
    dirAtributoIdAyuda:
      'El que no cambia nunca, que es lo que evita que un cambio de correo convierta a alguien en otra persona. En AD, objectGUID.',
    probarDirectorio: 'Probar la conexión',
    probando: 'Probando…',
    directorioOk: 'El directorio ha contestado y la cuenta de servicio entra.',
    keycloak: 'Keycloak',
    keycloakAyuda:
      'Lo que hace falta para entrar por Keycloak. El secreto del cliente se guarda y no se vuelve a enseñar.',
    kcInterno: 'Dirección interna del reino (opcional)',
    kcInternoAyuda:
      'La dirección que alcanza el servidor. Déjala vacía para usar el emisor público. En Docker: http://keycloak:8080/sso/realms/catalina-support.',
    kcEmisor: 'Emisor del reino',
    kcEmisorAyuda:
      'La dirección del reino, la que ve el navegador. Ejemplo: https://sso.empresa.com/realms/empresa.',
    kcCliente: 'Cliente',
    kcClienteAyuda: 'El identificador del cliente confidencial dado de alta en el reino.',
    kcSecreto: 'Secreto del cliente',
    kcSecretoPuesto: 'Hay un secreto guardado. Déjalo vacío para no cambiarlo.',
    kcSecretoVacio: 'No hay ningún secreto guardado.',
    kcVuelta: 'Dirección de vuelta',
    kcVueltaAyuda:
      'A donde vuelve Keycloak: la ruta de vuelta de esta aplicación. Tiene que estar dada de alta en el reino.',
    probarKeycloak: 'Probar la conexión',
    keycloakOk: 'El reino ha contestado y dice dónde está su pantalla de entrada.',
    numeracion: 'La numeración y el reparto',
    numeracionAyuda:
      'El prefijo de los números de ticket y cómo se reparten los que llegan. El idioma de la instalación llega con el alta de cuentas.',
    prefijo: 'Prefijo de los números',
    prefijoAyuda: 'De 2 a 8 caracteres, en mayúsculas y números. Ejemplo: CS.',
    prefijoAviso:
      'Cambiarlo no cambia los números ya emitidos: los tickets que existen conservan el suyo.',
    repartoPrincipal: 'Tickets principales',
    repartoInterno: 'Tickets internos',
    asignacion: 'Reparto',
    aviso: 'Aviso por correo',
    asignacionNinguna: 'No se reparten',
    asignacionPorTurnos: 'Por turnos',
    avisoANadie: 'A nadie',
    avisoATodoElEquipo: 'A todo el equipo',
    avisoAlAsignado: 'Al asignado',
    guardarNumeracion: 'Guardar la numeración y el reparto',
    numeracionGuardada: 'La numeración y el reparto están guardados.',
  },

  // La vista de primer arranque: lo que se pide **antes de que la instalación tenga puerta**
  // (`docs/primer-arranque.md`). La mayoría de los campos reutilizan las claves de `configuracion`:
  // son los mismos datos, y tenerlos escritos dos veces sería la forma segura de que digan cosas
  // distintas. Aquí sólo viven los textos propios del asistente.
  instalacion: {
    titulo: 'Primer arranque',
    intro:
      'Cinco pasos para dejar esta instalación en marcha. Cada paso se guarda al avanzar, así que puedes cerrar y seguir donde lo dejaste.',
    paso1: 'La instalación',
    paso2: 'Cómo se entra',
    paso3: 'Dónde está',
    paso4: 'El correo',
    paso5: 'La inteligencia artificial',
    pasoDeCuatro: 'Paso {paso} de {total}',
    siguiente: 'Siguiente',
    anterior: 'Anterior',
    terminar: 'Terminar la instalación',
    resumen: 'Resumen',
    motorResponde: 'El motor de IA responde: los dos resúmenes del ticket están disponibles.',
    // La contraseña de la cuenta de fábrica **no se pide aquí**: vive en el entorno, y contarlo es
    // justo lo que evita que alguien la busque en la pantalla (sección 4 del documento).
    cuentaDeFabrica:
      'La cuenta de fábrica se llama «admin» y su contraseña es ADMIN_PASSWORD, del archivo de entorno del servidor. Es la puerta que entra siempre, sea cual sea el método elegido, y aquí no se pide: no vive en la base de datos.',
    correoObligatorio:
      'El correo saliente es obligatorio para completar la instalación: se usa en altas, restablecimientos y avisos.',
    correoPuesto: 'Hay un correo saliente configurado.',
    credencialPuesta: 'Configurada',
    sinCredencial: 'No requerida',
    confirmacionExterna: 'Privacidad y posible costo',
    confirmacionAceptada: 'Aceptados',
    terminada: 'Esta instalación ya está terminada: no se configura dos veces.',
    correoHost: 'Servidor de correo',
    correoHostAyuda:
      'El nombre o la dirección del servidor saliente. Ejemplo: smtp.mi-institucion.org.',
    correoPuerto: 'Puerto',
    correoPuertoAyuda: '587 para el envío con TLS, 465 para TLS directo y 25 sin cifrar.',
    correoTls: 'Conexión cifrada (TLS)',
    correoTlsAyuda: 'En producción, sí: por ahí viajan las credenciales del correo.',
    correoUsuario: 'Usuario',
    correoUsuarioAyuda:
      'El usuario con el que se autentica el envío. Si el servidor no pide credenciales, se deja vacío.',
    correoContrasena: 'Contraseña',
    correoContrasenaPuesta: 'Hay una contraseña guardada. Déjalo vacío para no cambiarla.',
    correoContrasenaVacia: 'No hay ninguna contraseña guardada.',
    correoRemitenteNombre: 'Nombre del remitente',
    correoRemitenteCorreo: 'Correo del remitente',
    correoRemitenteCorreoAyuda: 'La dirección que verá quien reciba el correo.',
    probarCorreo: 'Probar la conexión',
    // Lo que hace y lo que **no** hace la prueba del correo, dicho en el propio mensaje.
    correoOk:
      'El servidor de correo ha contestado y la autenticación ha funcionado. No se ha enviado ningún correo.',
  },

  menu: {
    misTickets: 'Mis tickets',
    bandeja: 'Bandeja',
    ticketsPrincipales: 'Tickets principales',
    ticketsInternos: 'Tickets internos',
    categorias: 'Categorías y etiquetas',
    usuarios: 'Usuarios',
    configuracion: 'Configuración',
    opciones: 'Opciones',
    abrir: 'Abrir el menú',
    cerrar: 'Cerrar el menú',
    plegar: 'Plegar el menú',
    desplegar: 'Desplegar el menú',
  },

  // Los cuatro papeles, como se llaman en la interfaz. Los valores son los del backend.
  papeles: {
    usuario: 'Usuario',
    soporte: 'Soporte técnico',
    desarrollo: 'Desarrollo',
    administrador: 'Administrador',
  },

  // De dónde sale cada cuenta: es lo que decide si su contraseña es nuestra o de su directorio.
  // Son cortos a propósito: se leen como filtro encima de la lista y dentro de un desplegable.
  origenes: {
    local: 'Local',
    ad: 'Directorio (AD)',
    keycloak: 'Keycloak',
  },

  // Los estados del ticket, tal y como los llama el backend. Se leen en la bandeja, en la ficha y en
  // los filtros, siempre que quien mira trabaja con tickets.
  estados: {
    nuevo: 'Nuevo',
    'en progreso': 'En progreso',
    'en espera': 'En espera',
    escalado: 'Escalado',
    resuelto: 'Resuelto',
    cerrado: 'Cerrado',
  },

  // Y los mismos estados contados al usuario, que **no lee jerga interna**: para él `escalado` es
  // «En curso», porque no tiene por qué saber que ha pasado a Desarrollo (interfaz-y-experiencia, 3.5).
  estadosDelUsuario: {
    nuevo: 'Recibido',
    'en progreso': 'En curso',
    'en espera': 'Esperamos tu respuesta',
    escalado: 'En curso',
    resuelto: 'Resuelto',
    cerrado: 'Cerrado',
  },

  // Las frases del historial. Llevan nombres y estados dentro, y por eso viven enteras aquí y no se
  // montan a trozos en el código: en inglés el orden de las palabras cambia.
  historial: {
    creado: '{quien} abrió el ticket',
    creadoSistema: 'Se abrió el ticket',
    editado: '{quien} editó el ticket',
    asignado: '{quien} se lo asignó a {detalle}',
    estado: '{quien} lo pasó a «{estado}»',
    estadoSistema: 'El ticket volvió a «{estado}»',
    escalado: '{quien} lo escaló a Desarrollo',
    resuelto: '{quien} lo marcó como resuelto',
    cerrado: '{quien} lo cerró',
    reabierto: '{quien} lo reabrió',
    observador: '{quien} quitó a {detalle} de los observadores',
    observador_anadido: '{quien} añadió a {detalle} a los observadores',
    elSistema: 'El sistema',
  },

  usuarios: {
    titulo: 'Usuarios',
    descripcion: 'Las cuentas de la instalación. Desde aquí se dan de alta y se mantienen.',
    cargando: 'Cargando las cuentas…',

    // --- la lista ---
    nueva: 'Nueva cuenta',
    buscar: 'Buscar por nombre o correo',
    buscarEtiqueta: 'Buscar',
    filtroPapel: 'Papel',
    filtroOrigen: 'Origen',
    filtroEstado: 'Estado',
    todos: 'Todos',
    activas: 'Activas',
    desactivadas: 'Desactivadas',
    limpiar: 'Quitar los filtros',
    columnaCuenta: 'Cuenta',
    columnaPapel: 'Papel',
    columnaOrigen: 'Origen',
    columnaEstado: 'Estado',
    columnaUltimaEntrada: 'Última entrada',
    columnaAcciones: 'Acciones',
    activa: 'Activa',
    desactivada: 'Desactivada',
    nunca: 'Nunca ha entrado',
    editar: 'Editar',
    nombre: 'Nombre',
    apellidos: 'Apellidos',
    abrir: 'Abrir la ficha',
    desactivar: 'Desactivar',
    reactivar: 'Reactivar',
    reactivaSola:
      'Es una cuenta de Keycloak: se reactivará sola cuando esa persona entre por su camino, sin que nadie haga nada aquí.',
    mandarEnlace: 'Cambiar contraseña',
    sinResultados: 'No hay ninguna cuenta que coincida con esos filtros.',
    sinCuentas: 'Todavía no hay ninguna cuenta. Da de alta la primera.',
    total: 'cuentas',
    pagina: 'Página',
    anterior: 'Anterior',
    siguiente: 'Siguiente',
    tarjetaVer: 'Ver la ficha',

    // --- el alta ---
    altaTitulo: 'Nueva cuenta',
    altaAyuda:
      'La cuenta se crea aquí y, si es local, la persona recibe un enlace para establecer su contraseña.',
    correo: 'Correo electrónico',
    papel: 'Papel',
    origen: 'Origen',
    idioma: 'Idioma de los correos',
    origenNota:
      'Las cuentas de AD y Keycloak llevan el camino del directorio: no se dan de alta desde aquí. Quien está en el directorio entra con su cuenta y la suya se crea sola en ese primer acceso.',
    idiomaDeLaInstalacion: 'El de la instalación',
    crear: 'Crear la cuenta',
    creando: 'Creando…',
    creada: 'Cuenta creada. Se le ha mandado el enlace para establecer su contraseña.',
    creadaSinEnlace:
      'La cuenta está creada, pero el enlace de alta no ha salido. Mándalo otra vez desde la lista.',
    obligatorios: 'Hacen falta el nombre, los apellidos y el correo.',

    // --- lo que contestan las acciones ---
    mandarEnlacePregunta: '¿Mandar el enlace a {nombre}?',
    mandarEnlaceAlta:
      'Le llegará un correo para establecer su contraseña. El enlace caduca en 24 horas.',
    mandarEnlaceRecuperacion:
      'Le llegará un correo para poner una contraseña nueva: la que tenga ahora dejará de valer. El enlace caduca en 1 hora.',
    enlaceMandado: 'Enlace mandado otra vez. Caduca en 24 horas.',
    desactivadaHecha: 'Cuenta desactivada. Ya no puede entrar.',
    reactivadaHecha: 'Cuenta reactivada. Puede entrar otra vez.',
    guardada: 'Cambios guardados.',

    // --- la ficha ---
    fichaTitulo: 'La cuenta',
    datos: 'Los datos',
    estado: 'Estado',
    ultimaEntrada: 'Última entrada',
    contrasena: 'Contraseña',
    contrasenaPropia: 'La establece la persona desde el enlace que recibe.',
    contrasenaDirectorio: 'Es la de su directorio: no se cambia desde aquí.',
    sinCambios: 'No hay nada que guardar.',
    volver: 'Volver a la lista',
    acciones: 'Acciones',
    desactivarPregunta: '¿Desactivar esta cuenta?',
    desactivarConsecuencia:
      'Dejará de poder entrar en cuanto lo hagas, aunque tenga el navegador abierto: la siguiente petición suya no valdrá. No se borra nada, y sus tickets siguen contando su historia.',
    siDesactivar: 'Sí, desactivar',
    cambiarOrigen: 'Cambiar el origen',
    cambiarOrigenAyuda:
      'Pasar a local quita el vínculo con el directorio; si la cuenta se queda sin contraseña, se le manda el enlace de alta.',
    guardarOrigen: 'Guardar el origen',
    origenGuardado: 'Origen cambiado.',
    guardar: 'Guardar',
    cancelar: 'Cancelar',
  },

  perfil: {
    abrir: 'Mi perfil',
    titulo: 'Mi perfil',
    descripcion: 'Tu nombre, tus apellidos y el idioma en el que se te escriben los correos.',
    nombre: 'Nombre',
    apellidos: 'Apellidos',
    correo: 'Correo electrónico',
    correoAyuda: 'Tu correo es con lo que entras, y lo cambia un Administrador.',
    papel: 'Papel',
    papelAyuda: 'Tu papel lo cambia un Administrador.',
    idioma: 'Idioma de los correos',
    idiomaAyuda: 'Es el idioma de tus correos, y con el que se te enseña la aplicación.',
    guardar: 'Guardar',
    guardando: 'Guardando…',
    guardado: 'Tu perfil está guardado.',
    contrasena: 'Contraseña',
    cambiarContrasena: 'Cambiar mi contraseña',
  },

  // El editor de los correos, que vive en el módulo `mail` (docs/modules/mail.md).
  correos: {
    titulo: 'Los correos',
    descripcion:
      'Los diez correos que manda la aplicación, en los dos idiomas. El texto es la voz de la institución, y lo cambia quien manda.',
    queCorreo: 'Qué correo',
    editado: 'Este correo tiene tu texto, no el de fábrica.',
    deFabrica: 'Este correo lleva el texto de fábrica.',
    idiomaEs: 'Español',
    idiomaEn: 'English',
    asunto: 'Asunto',
    cuerpo: 'Cuerpo',
    cuerpoAyuda:
      'Se escribe con las herramientas de arriba sobre el texto seleccionado. También vale escribir el HTML a mano.',
    marcadoresDisponibles: 'Marcadores disponibles',
    marcadorImprescindible: 'imprescindible',
    faltaMarcador:
      'A este correo le falta un marcador que conviene que esté: {marcadores}. Se puede guardar así, pero mira si es lo que quieres.',
    vistaPrevia: 'Vista previa',
    sinVistaPrevia: 'Escribe algo para ver cómo queda.',
    barraDeFormato: 'Formato del texto',
    negrita: 'Negrita',
    cursiva: 'Cursiva',
    lista: 'Lista',
    enlace: 'Enlace',
    guardar: 'Guardar',
    guardando: 'Guardando…',
    guardado: 'Guardado. La vista previa ya enseña el texto nuevo.',
    prueba: 'Enviarme una prueba',
    pruebaEnviada: 'Prueba enviada a {correo}. Lleva el asunto con [prueba] delante.',
    restaurar: 'Volver al de fábrica',
    restaurarPregunta: '¿Volver al texto de fábrica?',
    restaurarConsecuencia:
      'El texto que hay ahora en ese idioma se pierde y vuelve el que trae la aplicación.',
    siRestaurar: 'Sí, volver',
    restaurarHecho: 'Este correo ha vuelto al texto de fábrica.',
    cancelar: 'Cancelar',
    cargando: 'Cargando los correos…',
  },

  // Los diez correos, como se llaman en la pantalla. Es una sección propia porque el tipo de los
  // textos admite un solo nivel de claves.
  plantillas: {
    'ticket.created': 'Ticket nuevo, al técnico',
    'ticket.escalated': 'Ticket escalado, a Desarrollo',
    'ticket.waitingUser': 'Le pedimos algo al usuario',
    'ticket.waitingSupport': 'Desarrollo necesita algo',
    'ticket.resolved': 'Ticket resuelto, al usuario',
    'ticket.closed': 'Ticket cerrado, al usuario',
    'ticket.backToSupport': 'El ticket vuelve a Soporte',
    'ticket.mentioned': 'Te han etiquetado en un ticket',
    'auth.invitation': 'Alta de una cuenta',
    'auth.recovery': 'Recuperar la contraseña',
    'auth.passwordChanged': 'Contraseña cambiada',
  },

  tickets: {
    // --- la bandeja ---
    nuevo: 'Nuevo ticket',
    buscar: 'Buscar por número, asunto o texto',
    buscarEtiqueta: 'Buscar',
    filtroEstado: 'Estado',
    filtroTipo: 'Tipo',
    tipoPrincipal: 'Principales',
    tipoInterno: 'Internos',
    tipoTodos: 'Todos',
    todosLosEstados: 'Todos',
    columnaNumero: 'Ticket',
    columnaAsunto: 'Asunto',
    columnaClasificacion: 'Clasificación',
    columnaEstado: 'Estado',
    columnaSolicitante: 'Solicitante',
    columnaResponsable: 'Responsable',
    columnaActualizado: 'Fecha actualización',
    columnaMotivo: 'Motivo',
    resumenes: 'De qué va el ticket',
    columnaUltimaAccion: 'Última acción',

    // Los estados de los dos campos que redacta el motor de IA (docs/modules/ai.md). **La pantalla
    // nunca enseña una clave**: si el motor no está o falló, lo dice en castellano.
    resumenPendiente: 'Generando…',
    resumenSinMotor: 'Sin motor de IA',
    resumenError: 'No se pudo resumir',
    resumenVacio: '—',
    iaRedaccionAbrir: 'Mejorar con IA',
    iaRedaccionTitulo: 'Mejorar la redacción con IA',
    iaRedaccionAyuda: 'Revisa el borrador y elige cómo debe sonar. La propuesta no se guardará hasta que la uses y publiques o guardes el editor principal.',
    iaRedaccionPrivacidad: 'El motor configurado procesará este texto. Si es un servidor o proveedor externo, el texto saldrá de esta instalación conforme a la configuración aceptada por el Administrador.',
    iaRedaccionBorrador: 'Borrador',
    iaRedaccionTono: 'Tonalidad',
    iaTonoProfesional: 'Profesional',
    iaTonoCordial: 'Cordial',
    iaTonoBreve: 'Breve',
    iaTonoEmpatico: 'Empática',
    iaTonoTecnico: 'Técnica',
    iaRedaccionMejorar: 'Mejorar',
    iaRedaccionRegenerar: 'Volver a generar',
    iaRedaccionUsar: 'Usar este texto',
    iaRedaccionGenerando: 'Mejorando la redacción…',
    iaRedaccionLista: 'La propuesta está lista para revisar.',
    // --- editar donde se muestra (decisión 74) ---
    editarElAsunto: 'Editar el asunto',
    editarLaDescripcion: 'Editar la descripción',
    anadirEtiquetaAlTicket: 'Añadir una etiqueta',
    anadirObservador: 'Añadir un observador',

    // --- las menciones y los observadores ---
    etiquetar: 'Etiquetar',
    etiquetarAyuda: 'Llama a otro técnico o desarrollador: pasa a seguir el ticket.',
    etiquetarA: 'A quién llamas',
    etiquetando: 'Etiquetando…',
    sinPersonas: 'No hay ningún técnico ni desarrollador activo al que llamar.',
    sinPersonasParaObservar: 'No hay ningún técnico ni desarrollador activo al que añadir.',
    observadores: 'Observadores',
    sinObservadores: 'Nadie más sigue este ticket.',
    quitarObservador: 'Quitar de los observadores',
    observadorQuitado: '{quien} ya no sigue el ticket.',

    // --- el chip de vista de «Mis tickets» ---
    vista: 'Vista',
    vistaAsignados: 'Asignados',
    vistaObservo: 'Observo',
    vistaTodos: 'Todos',

    // --- la categoría y las etiquetas (docs/interfaz-y-experiencia.md, sección 3.3) ---
    // La categoría es una palabra o dos y se lee de un vistazo; las etiquetas matizan. Las dos se
    // enseñan como chips debajo del número, y las dos filtran y buscan (docs/modules/tickets.md, 68).
    categoria: 'Categoría',
    etiquetas: 'Etiquetas',
    filtroCategoria: 'Categoría',
    filtroEtiqueta: 'Etiqueta',
    todasLasCategorias: 'Todas',
    buscarPorEtiqueta: 'Buscar por etiqueta',
    sinEtiquetas: 'Sin etiquetas',

    regenerar: 'Volver a resumir',
    regenerando: 'Pidiendo los resúmenes…',
    resumenesPedidos: 'Se han pedido otra vez los dos resúmenes.',
    sinResponsable: 'Sin responsable',
    cargando: 'Cargando los tickets…',
    vacia: 'No hay ningún ticket que coincida con esos filtros.',
    vaciaSinFiltros: 'Todavía no hay ningún ticket.',
    limpiar: 'Quitar los filtros',
    total: 'tickets',
    pagina: 'Página',
    anterior: 'Anterior',
    siguiente: 'Siguiente',
    verTicket: 'Abrir',
    verImagen: 'Ver la imagen en grande',
    verVideo: 'Ver el vídeo en grande',
    tarjetaAsignado: 'Lo lleva',

    // --- el alta ---
    nuevoTitulo: 'Nuevo ticket',
    nuevoAyuda: 'Cuenta qué pasa, con el mayor detalle que puedas. Si puedes, adjunta una captura.',
    asunto: 'Asunto',
    campoDescripcion: 'Descripción',
    camposDescripcionAyuda:
      'Escribe con formato y pon los archivos donde quieras que salgan: se ven aquí mismo.',
    archivoRechazado: 'Estos archivos no se han añadido:',
    paraQuien: '¿Para quién es el ticket?',
    paraQuienAyuda: 'Soporte puede crear un ticket en nombre de otra persona.',
    elegirSolicitante: 'La persona',
    crear: 'Crear el ticket',
    creando: 'Creando…',
    creado: 'Ticket creado: {numero}.',
    creadoConAdjuntos: 'Ticket creado: {numero}. Estos adjuntos no han subido:',
    obligatorios: 'Hacen falta el asunto y la descripción.',
    volver: 'Volver',
    volverALaBandeja: 'Volver a la bandeja',

    // El alta estrena la clasificación: **la categoría es obligatoria** —no hay ticket sin
    // clasificar— y las etiquetas son opcionales y se escriben normalizándose solas (decisiones 65 y 68).
    elegirCategoria: 'Elige una categoría',
    categoriaAyuda: 'Todo ticket va clasificado: Soporte y Desarrollo agrupan por aquí.',
    sinCategoriasActivas:
      'No hay ninguna categoría activa, así que ahora mismo no se puede crear un ticket. Pídele a Soporte o al Administrador que cree una.',
    etiquetasCampo: 'Etiquetas',
    guardandoEtiquetas: 'Guardando…',
    sinEtiquetasQueCoincidan: 'Ninguna etiqueta coincide con esa búsqueda.',
    editarEtiquetas: 'Editar las etiquetas',
    etiquetasAyuda:
      'Opcionales. En minúsculas y con guiones (red-wifi), hasta 32 caracteres; pulsa Intro o coma para añadirlas.',
    etiquetasSugeridas: 'Ya existen',
    anadirEtiqueta: 'Añadir la etiqueta',
    quitarEtiqueta: 'Quitar la etiqueta',

    // --- el detalle ---
    copiarNumero: 'Copiar el número',
    copiado: 'Copiado',
    copiarNoSePudo: 'No se ha podido copiar solo: selecciona el número y cópialo a mano.',
    conversacion: 'La conversación',
    ficha: 'La ficha',
    solicitante: 'Solicitante',
    creadoPor: 'Lo abrió',
    responsable: 'Responsable',
    abiertoEl: 'Abierto el',
    actualizadoEl: 'Última actualización',
    resueltoEl: 'Resuelto el',
    cerradoEl: 'Cerrado el',

    // La ficha enseña la clasificación y, para quien puede editar el ticket, la deja cambiar con el
    // mismo patrón que la asignación de responsable (docs/interfaz-y-experiencia.md, sección 3.4).
    cambiarCategoria: 'Cambiar la categoría',
    guardarClasificacion: 'Guardar la categoría y las etiquetas',
    clasificacionGuardada: 'Categoría y etiquetas guardadas.',
    sinEtiquetasFicha: 'Sin etiquetas.',
    motivoDelEscalado: 'Motivo del escalado',
    elPrincipal: 'El ticket principal',
    sinAdjuntos: 'Sin adjuntos',
    escribirComentario: 'Escribe un comentario',
    // Un comentario puede ser sólo un adjunto: se dice, porque si no, parece que hay que escribir. Y
    // las tres puertas para adjuntar se cuentan aquí, que es donde se escribe.
    comentarioAyuda:
      'Con formato, y con los archivos dentro: arrástralos aquí, pégalos con Ctrl+V o búscalos. Un comentario puede llevar sólo archivos.',
    enviar: 'Comentar',
    enviando: 'Enviando…',
    comentarioVacio: 'Escribe algo o adjunta un archivo antes de mandarlo.',
    comentarioEditado: 'editado',
    comentarioBorrado: 'Este comentario se borró. Su texto ya no está.',
    editar: 'Editar',
    borrar: 'Borrar',
    guardar: 'Guardar',
    cancelar: 'Cancelar',
    borrarPregunta: '¿Borrar este comentario?',
    borrarConsecuencia: 'El texto se vacía de verdad y queda la marca. Los adjuntos se quedan.',
    siBorrar: 'Sí, borrar',
    cargandoTicket: 'Cargando el ticket…',
    sinComentarios: 'Todavía no hay comentarios.',

    // --- la vista doble ---
    vistaPrincipal: 'Principal',
    vistaDoble: 'Los dos',
    vistaInterna: 'Interno',
    vistaEtiqueta: 'Qué quieres ver',

    // --- las acciones ---
    asignarA: 'Asignar a',
    asignar: 'Asignar',
    asignado: 'Ticket asignado.',
    empezar: 'Empezar',
    preguntarAlUsuario: 'Preguntar al usuario',
    resolver: 'Resolver',
    resolverTitulo: 'Resolver el ticket',
    resolverAyuda:
      'Cuéntale al usuario qué se ha hecho, en sus palabras: es lo único que va a leer.',
    resolviendo: 'Resolviendo…',
    cerrar: 'Cerrar',
    cerrarPregunta: '¿Cerrar este ticket?',
    estadoDelTicket: 'Estado del ticket',
    preguntarConsecuencia: 'El ticket queda esperando la respuesta del usuario, y se le avisa.',
    empezarConsecuencia: 'El ticket pasa a estar en progreso: queda en manos de quien lo trabaja.',
    pasarA: '¿Pasar el ticket a «{estado}»?',
    confirmar: 'Confirmar',
    comentarioDelCierre: 'Comentario del cierre',
    cierreAyuda: 'Queda en la conversación: es lo que se lee para saber por qué se cerró.',
    cerrarConsecuencia: 'Dejará de admitir comentarios. Se puede reabrir cuando haga falta.',
    siCerrar: 'Sí, cerrar',
    reabrir: 'Reabrir',
    reabrirAyuda: 'Este ticket está cerrado: para seguir hablando hay que reabrirlo.',
    accionesDelTicket: 'Acciones del ticket',
    queHacer: '¿Qué quieres hacer?',
    escalar: 'Escalar',
    escalarTitulo: 'Escalar',
    escalarAyuda:
      'Cuenta qué has probado, qué has descartado y qué hace falta. Sin el motivo, esto es un «pásalo tú».',
    motivo: 'Motivo',
    escalarBoton: 'Escalar',
    // Lo que pide cada acción en su diálogo, con su propio nombre: nada de dos campos que se llamen
    // igual en la misma pantalla.
    laPregunta: 'La pregunta',
    enviarLaPregunta: 'Enviar la pregunta',
    loQueNecesitas: 'Qué necesitas de Soporte',
    pedirlo: 'Pedirlo',
    queSeHizo: 'Qué se ha hecho',
    resolverYAvisar: 'Resolver y avisar',
    porQueNoEsCodigo: 'Por qué no es un cambio de código',
    escalando: 'Escalando…',
    pedirAlgo: 'Necesito algo de Soporte',
    devolver: 'No es un cambio de código',
    devolverPregunta: '¿Devolver el caso a Soporte?',
    devolverConsecuencia:
      'El ticket principal vuelve a la bandeja de Soporte, que es quien habla con el usuario, y se le avisa.',
    siDevolver: 'Sí, devolver',
    sinAcciones: 'No hay nada que hacer aquí: este ticket se mira.',

    // --- lo que contestan las acciones ---
    ticketActualizado: 'Ticket actualizado.',
    categoriaCambiada: 'Categoría cambiada.',
    etiquetasCambiadas: 'Etiquetas guardadas.',
    observadorAnadido: 'Observador añadido.',
    descripcionActualizada: 'Descripción actualizada.',
    estadoCambiado: 'El ticket ha cambiado de estado.',
    resueltoAviso: 'Ticket resuelto. Se ha avisado al usuario.',
    cerradoAviso: 'Ticket cerrado.',
    reabiertoAviso: 'Ticket reabierto.',
    escaladoAviso: 'Ticket escalado. Se ha avisado a Desarrollo.',
    comentarioEscrito: 'Comentario escrito.',
    // El comentario ya está publicado y algún archivo no ha subido: se dice cuál y se puede reintentar
    // sin volver a escribirlo, que es lo que evita el comentario duplicado.
    adjuntosQueFaltan: 'El comentario se ha publicado, pero no se han podido subir estos archivos:',
    reintentarAdjuntos: 'Reintentar la subida',
    adjuntosSubidos: 'Ya están subidos todos los archivos del comentario.',
    comentarioGuardado: 'Comentario guardado.',
    comentarioBorradoAviso: 'Comentario borrado.',
    adjuntoSubido: 'Adjunto subido.',
    archivoNoTraido: 'No se ha podido traer el archivo.',
    // El hueco de un archivo que se nombró en el texto y no llegó a subir (decisión 44): se ve dónde
    // estaba, y no como un hueco mudo que parezca un fallo de la pantalla.
    adjuntoNoSubio: 'No subió',
    descargar: 'Descargar',
    abrirEnPestana: 'Abrir en una pestaña',
    ver: 'Ver',
    vistaPrevia: 'Vista previa',
    cerrarVistaPrevia: 'Cerrar la vista previa',
    // Lo que se enseña mientras se está trayendo un adjunto para pintarlo dentro del texto.
    trayendoAdjunto: 'Trayendo el archivo…',

    // --- la búsqueda con modal: todos los criterios en un sitio, y un resumen en la lista ---
    buscarTitulo: 'Buscar tickets',
    abrirBusqueda: 'Buscar',
    buscarAyuda: 'Se pueden combinar todos los criterios: lo que salga los cumple todos.',
    resumenFiltros: '{n} filtros',
    resumenUnFiltro: '1 filtro',
    sinFiltros: 'Sin filtros',
    quitarLosFiltros: 'Quitar los filtros',

    // --- las etiquetas se mantienen, como las categorías ---
    nuevaEtiqueta: 'Nueva etiqueta',
    creandoEtiqueta: 'Creando…',
    crearEtiqueta: 'Crear',
    renombrarEtiqueta: 'Renombrar',
    retirarEtiqueta: 'Retirar',
    etiquetaCreada: 'Etiqueta creada.',
    etiquetaRenombrada: 'Etiqueta renombrada en todos los tickets que la llevan.',
    etiquetaRetirada: 'Etiqueta retirada y quitada de sus tickets.',
    avisoDeRetiradaDeEtiqueta:
      'Se va a retirar «{etiqueta}» y a quitarla de {n} tickets. Se puede volver a escribir en cualquier ticket.',
    etiquetaSinTickets: 'Todavía no la lleva ningún ticket.',
    nombreDeEtiqueta: 'Etiqueta',
    etiquetasDescripcion:
      'Con esto se matiza lo que una categoría no distingue. Nacen solas al escribirlas en un ticket, y aquí se corrigen y se retiran.',

    // --- «Categorías y etiquetas»: el catálogo (docs/interfaz-y-experiencia.md, sección 3.7) ---
    // La lista de categorías la mantienen Soporte y el Administrador; **retirar es sólo del
    // Administrador** (decisión 64). Abajo, las etiquetas que se están usando, con su cuenta.
    categoriasTitulo: 'Categorías y etiquetas',
    categoriasDescripcion:
      'Con esto se clasifica cada ticket. Retirar una categoría deja de ofrecerla al crear, pero no toca los tickets que ya la tienen.',
    categoriasLista: 'Categorías',
    nuevaCategoria: 'Nueva categoría',
    nombreDeCategoria: 'Nombre',
    crearCategoria: 'Crear',
    creandoCategoria: 'Creando…',
    renombrarCategoria: 'Renombrar',
    retirarCategoria: 'Retirar',
    reactivarCategoria: 'Volver a poner',
    categoriaActiva: 'Activa',
    categoriaRetirada: 'Retirada',
    sinCategorias: 'Todavía no hay ninguna categoría.',
    sinEtiquetasEnUso: 'Todavía no hay ninguna etiqueta en uso.',
    categoriaCreada: 'Categoría creada.',
    categoriaRenombrada: 'Categoría renombrada.',
    categoriaRetiradaAviso: 'Categoría retirada. Los tickets que la tienen la conservan.',
    categoriaReactivadaAviso: 'Categoría puesta otra vez.',
  },

  // Los botones del editor con formato, que es el cuadro de escribir de un ticket y de un comentario
  // (docs/interfaz-y-experiencia.md, sección 6.5). Es una sección propia, como `historial`: el tipo de
  // los textos admite un solo nivel de claves.
  editor: {
    barra: 'Formato del texto',
    negrita: 'Negrita',
    cursiva: 'Cursiva',
    subrayado: 'Subrayado',
    tachado: 'Tachado',
    listaVinietas: 'Lista con viñetas',
    listaNumerada: 'Lista numerada',
    enlace: 'Enlace',
    adjuntar: 'Adjuntar',
    direccion: 'La dirección del enlace',
    direccionPoner: 'Poner el enlace',
    direccionCancelar: 'Cancelar',
    // El enlace se comprueba antes de ponerlo: lo que no empiece por uno de los tres esquemas se
    // rechaza, y es la misma regla que aplica el backend al guardar (docs/modules/tickets.md, 2.3).
    direccionInvalida: 'La dirección tiene que empezar por http, https o mailto.',
  },

  tema: {
    etiqueta: 'Tema',
    claro: 'Claro',
    oscuro: 'Oscuro',
    papel: 'Papel',
    niebla: 'Niebla',
    contraste: 'Alto contraste',
    grafito: 'Grafito',
    noche: 'Noche',
    sepia: 'Sepia',
    // Los dos grupos en los que se ordenan los ocho.
    claros: 'Claros',
    oscuros: 'Oscuros',
  },

  entrada: {
    titulo: 'Entrar',
    descripcion: 'Catalina Support',
    correo: 'Correo electrónico',
    correoAyuda: 'Escribe el correo con el que te dieron de alta.',
    contrasena: 'Contraseña',
    entrar: 'Entrar',
    entrando: 'Entrando…',
    olvidada: 'He olvidado mi contraseña',
    obligatorios: 'Escribe tu correo y tu contraseña.',
    // El camino de Keycloak: el botón se enseña **sólo si esta instalación lo tiene**, y lo dice el
    // backend (docs/modules/auth.md, decisión 27).
    oCon: 'O entra con la cuenta de tu organización',
    conKeycloak: 'Entrar con Keycloak',
    // Con el método en AD el formulario **es el mismo** —correo y contraseña—, pero la contraseña no
    // es de aquí: se dice, y se quita el enlace de recuperarla, que no serviría de nada
    // (docs/modules/settings.md, sección 5.8).
    porDirectorio:
      'Esta instalación entra con las cuentas de la organización: escribe tu correo y tu contraseña de siempre.',
    soloKeycloak: 'Esta instalación entra con Keycloak: se entra desde su pantalla, no desde aquí.',
    // La puerta de la cuenta de fábrica: con el método en Keycloak no hay formulario, y **sin ella la
    // instalación se quedaría sin nadie que pudiera volver a cambiar el método**
    // (docs/usuarios-y-permisos.md, sección 8).
    comoAdministrador: 'Entrar como administrador',
    cuentaDeFabrica:
      'La contraseña de la cuenta de fábrica vive en la configuración de la instalación.',
  },

  olvido: {
    titulo: 'Recuperar la contraseña',
    explicacion:
      'Escribe tu correo y te mandamos un enlace para establecer una contraseña nueva. El enlace caduca en 1 hora.',
    correo: 'Correo electrónico',
    enviar: 'Mandar el enlace',
    enviando: 'Mandando…',
    enviado:
      'Si esa dirección tiene cuenta, te ha llegado un enlace. Revisa tu correo, y si no aparece, mira en la carpeta de spam.',
    volver: 'Volver a la pantalla de entrada',
  },

  establecer: {
    titulo: 'Establece tu contraseña',
    explicacion: 'Escribe la contraseña que quieras usar, dos veces.',
    nueva: 'Contraseña nueva',
    repetir: 'Repite la contraseña',
    guardar: 'Guardar la contraseña',
    guardando: 'Guardando…',
    ayuda: 'Mínimo 8 caracteres.',
    sinToken:
      'Este enlace no trae la información necesaria. Pide otro desde la pantalla de entrada.',
    noCoinciden: 'Las dos contraseñas no son iguales.',
    hecha: 'Tu contraseña ya está puesta. Entra con ella.',
  },

  cambio: {
    titulo: 'Cambiar mi contraseña',
    actual: 'Contraseña actual',
    nueva: 'Contraseña nueva',
    repetir: 'Repite la contraseña nueva',
    guardar: 'Cambiar la contraseña',
    guardando: 'Cambiando…',
    ayuda: 'Mínimo 8 caracteres.',
    noCoinciden: 'Las dos contraseñas nuevas no son iguales.',
    hecha: 'Tu contraseña ha cambiado. Te hemos mandado un aviso por correo.',
  },

  sinPermiso: {
    titulo: 'No puedes ver esta pantalla',
    explicacion:
      'Tu cuenta no tiene permiso para lo que has pedido. Si crees que es un error, habla con Soporte.',
    volver: 'Volver a lo mío',
  },

  servidor: {
    titulo: 'No hay conexión con el servidor',
    explicacion:
      'No hemos podido hablar con el servidor de {instalacion}. Puede ser un corte momentáneo: vuelve a intentarlo en unos segundos.',
    reintentar: 'Reintentar',
  },

  sesion: {
    salir: 'Salir',
    salirPregunta: '¿Salir de la aplicación?',
    salirConsecuencia: 'Tendrás que volver a entrar con tu cuenta.',
    caducada: 'Tu sesión ha caducado. Entra otra vez.',
  },

  comun: {
    volver: 'Volver',
    cerrar: 'Cerrar',
    cancelar: 'Cancelar',
    cargando: 'Cargando…',
    obligatorio: 'Este campo es obligatorio.',
  },

  /*
   * Las claves de error del backend, con su texto. Están **todas** las que el backend puede
   * mandar: si se añade una clave nueva allí, aquí falta y no compila.
   */
  errores: {
    'error interno': 'Algo ha fallado por nuestra parte. Vuelve a intentarlo en un momento.',

    // El motor de IA (docs/modules/ai.md): o no está, o contestó algo que no vale.
    'ai.unavailable': 'El motor de resúmenes no está disponible. Los tickets funcionan igual.',
    'ai.invalid': 'El motor de resúmenes contestó algo que no se entiende.',
    'mail.translation.costConfirmationRequired': 'Confirma el posible costo del proveedor o edita las plantillas manualmente.',
    'mail.translation.incomplete': 'Revisa las once plantillas antes de aplicar el idioma.',
    'mail.translation.protectedChanged': 'La traducción alteró el HTML o un marcador. Corrige esa plantilla manualmente.',
    'settings.language.reviewRequired': 'Revisa las once plantillas antes de aplicar el idioma.',
    'tickets.mention.notAllowed': 'No puedes etiquetar a nadie en este ticket.',

    'auth.invalidCredentials': 'El correo o la contraseña no son correctos.',
    'auth.accountInactive': 'Tu cuenta está desactivada. Habla con Soporte para que la activen.',
    'auth.session.invalid': 'Tu sesión no vale. Entra otra vez.',
    'auth.session.expired': 'Tu sesión ha caducado. Entra otra vez.',
    'auth.forbidden': 'Tu cuenta no tiene permiso para eso.',
    'auth.token.expired': 'Ese enlace ha caducado. Pide otro.',
    'auth.token.used':
      'Ese enlace ya no vale: puede que se haya usado ya, o que lo sustituyera otro más nuevo.',
    'auth.password.tooShort': 'La contraseña tiene que tener al menos 8 caracteres.',
    'auth.password.wrong': 'La contraseña actual no es la que has escrito.',
    'auth.password.notLocal':
      'Esa contraseña no se cambia desde aquí: es la de tu directorio de trabajo.',
    'auth.oidc.state': 'La vuelta del acceso no vale o ha caducado. Prueba a entrar otra vez.',
    'auth.oidc.notConfigured': 'Esta instalación no tiene Keycloak configurado.',
    'auth.oidc.rejected': 'No se ha podido completar la entrada con Keycloak.',
    'auth.oidc.unavailable':
      'Keycloak no responde. Inténtalo en un momento o entra con tu correo y tu contraseña.',
    'auth.oidc.noEmail':
      'Tu cuenta de la organización no tiene correo electrónico, y sin correo no se puede crear la cuenta aquí. Avisa a Soporte.',
    'auth.directory.unavailable':
      'El directorio de tu organización no responde. Inténtalo en un momento.',
    'auth.directory.rejected': 'El directorio ha rechazado esas credenciales.',

    'settings.name.tooLong': 'El nombre no puede pasar de 60 caracteres.',
    'settings.language.unknown': 'Ese idioma no existe.',
    'settings.primaryColor.invalid':
      'Ese color no vale: tiene que ser un color en hexadecimal, como #1d4ed8.',
    'settings.numberPrefix.invalid':
      'El prefijo tiene que ser de 2 a 8 caracteres, en mayúsculas y números.',
    'settings.assignment.unknown': 'Esa forma de repartir no existe.',
    'settings.notification.unknown': 'Esa forma de avisar no existe.',
    'settings.notification.withoutAssignment':
      'No se puede avisar sólo al asignado si el ticket no se reparte: no habría a quién avisar.',
    'settings.logo.format': 'Ese archivo no es una imagen de las que valen: PNG, JPEG, WebP o SVG.',
    'settings.logo.invalid':
      'Esa imagen no se ha podido leer, o es más grande de lo que se admite.',
    'settings.logo.tooBig': 'Ese logo pesa demasiado: el máximo es 1 MB.',
    'settings.logo.variantUnknown': 'Ese hueco de logo no existe: es «claro» u «oscuro».',
    'settings.logo.notFound':
      'La instalación no tiene un logo propio: se está usando el de fábrica.',
    'settings.notFound': 'No se ha encontrado la configuración de la instalación.',
    'settings.method.unknown': 'Ese método de entrada no existe: es «local», «ad» o «keycloak».',
    'settings.method.notConfigured':
      'Ese método todavía no está configurado, y elegirlo dejaría la instalación sin puerta para todo el mundo menos la cuenta de fábrica. Configúralo primero.',
    'settings.directory.incomplete': 'Al directorio le falta el servidor o la base de búsqueda.',
    'settings.keycloak.incomplete':
      'A Keycloak le falta el emisor, el cliente o la dirección de vuelta.',
    'settings.directory.unreachable':
      'No se ha podido conectar con el directorio: revisa el servidor, el puerto y la cuenta de servicio.',
    'settings.keycloak.unreachable':
      'No se ha podido leer el reino de Keycloak: revisa el emisor y que el reino exista.',

    'users.notFound': 'Esa cuenta no existe.',
    'users.email.duplicated': 'Ese correo ya está en uso por otra cuenta.',
    'users.email.invalid': 'Ese correo no tiene forma de correo.',
    'users.name.required': 'Faltan el nombre o los apellidos.',
    'users.role.notAllowed': 'No puedes repartir ese papel.',
    'users.origin.unknown': 'Ese origen de cuenta no existe.',
    'users.directory.notFound': 'El directorio no conoce a esa persona: ya no está allí.',
    'users.origin.byDirectory':
      'Las cuentas del directorio no se crean ni se cambian de origen a mano: quien está en el directorio entra con su cuenta y la suya se pone al día en ese acceso.',
    'users.directory.activatesItself':
      'Esa cuenta entra sola por Keycloak: en cuanto esa persona entre, volverá a estar activa sin que nadie haga nada.',
    'users.directory.unavailable':
      'No se ha podido comprobar con el directorio. Inténtalo en un momento.',
    'users.origin.inUse': 'No se puede cambiar el origen de una cuenta de directorio activa.',
    'users.password.notLocal':
      'La contraseña de esa cuenta la comprueba el directorio: no se cambia desde aquí.',
    'settings.timeZone.unknown': 'Esa zona horaria no existe. Elige una de la lista.',
    'settings.publicUrl.invalid':
      'La dirección tiene que empezar por http:// o https:// y llevar su servidor.',
    'settings.aiUrl.invalid':
      'La dirección del motor de IA no vale: no hay ninguna que probar, o no empieza por http:// o https:// y le falta su servidor.',
    'settings.ai.unreachable':
      'El motor de IA no ha contestado: revisa la dirección y que esté levantado.',
    'settings.ai.managerUnavailable': 'El administrador local de modelos no está disponible.',
    'settings.ai.insufficientMemory':
      'El modelo requiere {requerida} de RAM y hay {disponible} disponibles. El modelo anterior continúa activo.',
    'settings.ai.required': 'Configura y prueba el motor de IA antes de entrar al producto.',
    'settings.unavailable': 'No se ha podido leer la configuración de la instalación.',
    'settings.ai.licenseRequired': 'Debes aceptar la licencia antes de descargar el modelo.',
    // Las cuatro claves del asistente de primer arranque (docs/primer-arranque.md, sección 6).
    'setup.alreadyInstalled': 'Esta instalación ya está terminada: no se configura dos veces.',
    'setup.step.incomplete': 'A este paso le falta algún dato. Repásalo y vuelve a intentarlo.',
    'setup.mail.incomplete':
      'Para el correo hacen falta el servidor, el puerto y la dirección del remitente.',
    'setup.mail.unreachable':
      'El servidor de correo no ha contestado o no ha aceptado el usuario y la contraseña.',
    'auth.publicUrl.missing':
      'No hay dirección pública configurada: sin ella no se puede armar el enlace que va en el correo. Se configura en Configuración.',
    'users.accountInactive': 'Esa cuenta está desactivada: actívala antes de mandarle un enlace.',
    'users.selfDeactivation': 'No puedes desactivar tu propia cuenta: pídeselo a otra persona.',
    // Un aviso, no un error: la acción se hace, pero conviene saber esto.
    'users.directoryMayReturn':
      'Es una cuenta de directorio: si allí sigue activa, volverá a entrar y se reactivará sola.',
    'users.stateHasItsOwnAction':
      'El estado de una cuenta se cambia con su propia acción, no editando la cuenta.',

    'tickets.notFound': 'Ese ticket no existe.',
    'tickets.subject.required': 'Escribe un asunto.',
    'tickets.description.required': 'Escribe una descripción.',
    'tickets.writing.draft.required': 'Escribe un borrador antes de pedir una mejora.',
    'tickets.writing.editor.invalid': 'Ese editor no admite la mejora de redacción.',
    'tickets.writing.tone.invalid': 'Elige una tonalidad válida.',
    'tickets.writing.unavailable': 'El motor de IA no está disponible. El borrador se conserva.',
    'tickets.writing.invalid': 'El motor devolvió una propuesta que no se puede utilizar.',
    // El saneador del backend sólo admite las etiquetas y los atributos de su lista blanca, y lo que
    // no está **no se guarda**: se dice, porque un texto que se guarda a medias es peor que uno que se
    // rechaza (docs/modules/tickets.md, sección 2.3). Con el pegado sin formato del editor, a lo que
    // llega aquí se le quita el formato de lo que se había pegado.
    'tickets.body.notAllowed':
      'El texto lleva algo que no se admite. Quita el formato de lo que has pegado y vuelve a intentarlo.',
    'tickets.reason.required':
      'Escribe por qué se escala: sin el motivo, Desarrollo no sabe qué hace falta.',
    'tickets.comment.required': 'Escribe un comentario.',
    'tickets.state.unknown': 'Ese estado no existe.',
    'tickets.transition.notAllowed': 'Ese cambio de estado no se puede hacer así.',
    'tickets.forbidden': 'Eso no te toca en este ticket.',
    'tickets.closed': 'El ticket está cerrado: reábrelo para seguir.',
    'tickets.comment.notYours': 'Sólo quien escribió un comentario puede cambiarlo.',
    'tickets.comment.notFound': 'Ese comentario no existe.',
    'tickets.assignee.unknown': 'Esa persona no puede ser la responsable de este ticket.',
    'tickets.requester.unknown': 'Esa persona no existe o está desactivada.',
    'tickets.requester.notAllowed': 'Sólo Soporte puede crear un ticket en nombre de otra persona.',
    'tickets.attachment.extension': 'Ese tipo de archivo no se admite.',
    'tickets.attachment.tooBig': 'Ese archivo pesa demasiado: el máximo son 25 MB.',
    'tickets.attachment.invalid': 'No se ha podido leer el archivo.',
    'tickets.attachment.notFound': 'Ese adjunto no existe.',

    // Las claves de la categoría y de la etiqueta (docs/modules/tickets.md, sección 2.3.2).
    'tickets.category.required': 'Hace falta elegir una categoría: no hay ticket sin clasificar.',
    'tickets.category.notFound': 'Esa categoría no existe.',
    'tickets.category.inactive':
      'Esa categoría está retirada: elige otra para clasificar el ticket.',
    'tickets.category.name.required': 'Escribe un nombre para la categoría.',
    'tickets.category.duplicate': 'Ya hay una categoría con ese nombre.',
    'tickets.category.lastActive':
      'No se puede retirar la última categoría activa: el catálogo no se puede quedar sin ninguna.',
    'tickets.tag.forbidden': 'Sólo el Administrador puede retirar una etiqueta.',
    'tickets.tag.duplicate': 'Esa etiqueta ya existe.',
    'tickets.cierre.sinComentario': 'Para cerrar el ticket hay que decir por qué.',
    'tickets.etiqueta.desconocida':
      'Esa etiqueta no está en el catálogo. Sólo el Administrador puede crear etiquetas nuevas.',
    'tickets.tag.notFound': 'Esa etiqueta no existe.',
    'tickets.category.forbidden':
      'Sólo el Administrador puede retirar o volver a poner una categoría.',
    'tickets.tag.required': 'Escribe algo para la etiqueta: no puede quedar vacía.',
    // El tope lo impide el propio campo, pero si algo llega de otra puerta se dice igual.
    'tickets.tag.tooLong': 'La etiqueta no puede pasar de 32 caracteres.',

    'mail.template.notFound': 'Esa plantilla no existe.',
    'mail.template.unknownKey': 'Ese correo no es ninguno de los diez.',
    'mail.language.unknown': 'Ese idioma no existe.',
    'mail.subject.required': 'El asunto no puede estar vacío.',
    'mail.body.required': 'El cuerpo no puede estar vacío.',
    'mail.marker.unknown': 'El texto usa un marcador que ese correo no admite.',
    'mail.marker.missing': 'Falta el valor de un marcador que el texto usa.',
    'mail.test.noEmail':
      'La cuenta de fábrica no tiene correo, así que no hay dónde mandar la prueba.',
    'mail.smtp.notConfigured': 'No hay servidor de correo configurado.',
    'mail.smtp.connect': 'No se ha podido conectar con el servidor de correo.',
    'mail.smtp.handshake': 'El servidor de correo no ha aceptado la conexión cifrada.',
    'mail.smtp.auth': 'El servidor de correo ha rechazado las credenciales.',
    'mail.smtp.from': 'El servidor de correo ha rechazado el remitente.',
    'mail.smtp.recipient': 'El servidor de correo ha rechazado el destinatario.',
    'mail.smtp.data': 'El servidor de correo ha cortado el envío.',
    'mail.smtp.write': 'No se ha podido escribir el correo al servidor.',
    'mail.smtp.close': 'El servidor de correo ha cerrado la conexión antes de terminar.',
    'mail.to.empty': 'No hay destinatarios.',
    'mail.to.invalid': 'Alguna dirección de destinatario no vale.',
    'mail.from.invalid': 'La dirección del remitente no vale.',
  },
} as const;

/** Las claves de error que el backend puede mandar. Se sacan de la lista de arriba, no se copian. */
export type ClaveError = keyof typeof ES.errores;

/** Todos los textos de la interfaz. El inglés está obligado a tener estas mismas claves. */
export type Textos = {
  readonly [seccion in keyof typeof ES]: {
    readonly [clave in keyof (typeof ES)[seccion]]: string;
  };
};
