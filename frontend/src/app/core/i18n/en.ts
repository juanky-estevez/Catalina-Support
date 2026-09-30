import type { Textos } from './es';

/**
 * Los mismos textos, en inglés.
 *
 * Está tipado con `Textos`, que sale del diccionario español: si aquí falta una clave, o sobra
 * una que allí no está, **no compila**. Es lo que hace imposible que una pantalla enseñe una
 * clave sin traducir (docs/modules/auth.md, decisión 26).
 */
export const EN: Textos = {
  idioma: {
    nombre: 'English',
    cambiar: 'Change language',
    es: 'Español',
    en: 'English',
  },

  configuracion: {
    titulo: 'Settings',
    nombre: 'The installation name',
    nombreAyuda:
      'The institution, the company or the team. It is what is read on the sign-in screen, in the menu and in the browser tab.',
    nombreCampo: 'Name',
    nombreDeFabrica: 'Leave it empty and the factory name comes back (Catalina Support).',
    guardarNombre: 'Save the name',
    nombreGuardado: 'Name saved. It is already visible across the application.',
    marca: 'The logo',
    marcaAyuda: 'The logo of the installation. It is shown on the sign-in screen and in the menu.',
    logoClaro: 'Logo for light themes',
    logoClaroAyuda: 'Shown in Light, Paper, Mist and High contrast.',
    logoOscuro: 'Logo for dark themes',
    logoOscuroAyuda:
      'Shown in Dark, Graphite, Night and Sepia. If you do not set one, the other is used.',
    vistaPrevia: 'Preview',
    logoPuesto: 'You have a logo of your own in place.',
    logoDeFabrica: 'The factory logo is being used.',
    elegirArchivo: 'Choose a file',
    formatos: 'PNG, JPEG, WebP or SVG, up to 1 MB and 2000 pixels per side.',
    subir: 'Upload the logo',
    subiendo: 'Uploading…',
    volverAlDeFabrica: 'Back to the factory one',
    logoSubido: 'Logo uploaded. It is already in use.',
    logoQuitado: 'That slot is back to the factory logo.',
    color: 'The institutional colour',
    colorAyuda:
      'The colour of the house. It applies to the two factory themes; the other themes carry their own.',
    colorElegido: 'Colour',
    guardarColor: 'Save the colour',
    guardando: 'Saving…',
    descartar: 'Discard',
    colorGuardado: 'Colour saved.',
    ejemploBoton: 'A sample button',
    ejemploTexto: 'and some text next to it, to see how it all looks.',
    colorResuelto: 'On light themes it becomes',
    colorResueltoOscuro: 'and on dark ones',
    idioma: 'The installation language',
    idiomaAyuda:
      'The language new accounts are born with when whoever creates them does not choose another, and the language of the emails written to them.',
    correos: 'The emails',
    correosAyuda:
      'The ten emails the application sends, in both languages, are edited on their own screen: it belongs to the mail module, and from here you only get to it.',
    editarCorreos: 'Edit the emails',
    region: 'Time zone',
    regionAyuda:
      'The zone every date is read in, here and in the e-mails. They are stored in UTC: changing it moves no ticket.',
    buscarZona: 'Search a city or a zone',
    sinZonas: 'No zone matches what you are looking for.',
    horaDeLaZona: 'It is now {hora} ({desfase}).',
    regionYDireccion: 'Time zone and public address',
    regionYDireccionAyuda:
      'The zone dates are read in and the address this installation is reached at.',
    guardarRegion: 'Save the time zone and the address',
    regionGuardada: 'The time zone and the address have been saved.',
    direccion: 'Public address',
    direccionAyuda:
      'The address this installation is reached at: it is the base of the links in the e-mails and of Keycloak\'s return. http or https, with a port if needed, and localhost too.',
    direccionEjemplo: 'https://support.my-institution.org',
    direccionInsegura:
      'This installation is not served over https: the password and the session travel unencrypted over the network. It can be used like this —to try it locally, for instance—, and for real use a certificate is worth having.',
    entrada: 'How people sign in',
    entradaAyuda:
      'The installation signs people in through one method at a time: the other two stay off even if they are configured. The factory account always gets in, and from it the choice can be changed.',
    metodo: 'Sign-in method',
    metodoLocal: 'Application accounts',
    metodoLocalAyuda: 'Accounts live here and people sign in with an email address and a password.',
    metodoAD: 'Organization directory (AD)',
    metodoADAyuda:
      'People sign in with their organization account, against the directory. Local passwords stop working, even for Support and Development.',
    metodoKeycloak: 'Keycloak',
    metodoKeycloakAyuda:
      'People sign in from the Keycloak screen, with their account. There is no email and password form here.',
    metodoSinConfigurar: 'Configure it below before you can choose it.',
    guardarEntrada: 'Save the sign-in method',
    entradaGuardada: 'Saved. It applies from the next sign-in, with no restart.',
    directorio: 'The organization directory',
    directorioAyuda:
      'What is needed to sign in through AD. The service account password is stored and never shown again: leave it empty to keep the current one.',
    dirServidor: 'Server',
    dirServidorAyuda:
      'The name or address of the directory. For example: ldap or ad.company.local.',
    dirPuerto: 'Port',
    dirPuertoAyuda: '389 for LDAP and 636 for LDAPS. Leave it empty for 389.',
    dirTls: 'Encrypted connection (LDAPS)',
    dirTlsAyuda: 'In production, yes: credentials travel through it.',
    dirCuenta: 'Service account',
    dirCuentaAyuda:
      'The full name used to look people up. For example: cn=lookup,ou=services,dc=company,dc=local.',
    dirContrasena: 'Service account password',
    dirContrasenaPuesta: 'A password is stored. Leave it empty to keep it.',
    dirContrasenaVacia: 'No password is stored.',
    dirBase: 'Where to search',
    dirBaseAyuda: 'The search base. For example: ou=people,dc=company,dc=local.',
    dirFiltro: 'Search filter',
    dirFiltroAyuda:
      'With “%s” where the email address goes. For example: (mail=%s) or (userPrincipalName=%s).',
    dirAtributoCorreo: 'Email attribute',
    dirAtributoNombre: 'First name attribute',
    dirAtributoApellidos: 'Last name attribute',
    dirAtributoId: 'Identifier attribute',
    dirAtributoIdAyuda:
      'The one that never changes, which is what stops an email address change from turning someone into somebody else. In AD, objectGUID.',
    probarDirectorio: 'Test the connection',
    probando: 'Testing…',
    directorioOk: 'The directory answered and the service account gets in.',
    keycloak: 'Keycloak',
    keycloakAyuda:
      'What is needed to sign in through Keycloak. The client secret is stored and never shown again.',
    kcEmisor: 'Realm issuer',
    kcEmisorAyuda:
      'The realm address, the one the browser sees. For example: https://sso.company.com/realms/company.',
    kcCliente: 'Client',
    kcClienteAyuda: 'The identifier of the confidential client registered in the realm.',
    kcSecreto: 'Client secret',
    kcSecretoPuesto: 'A secret is stored. Leave it empty to keep it.',
    kcSecretoVacio: 'No secret is stored.',
    kcVuelta: 'Redirect address',
    kcVueltaAyuda:
      'Where Keycloak comes back to: this application’s callback route. It has to be registered in the realm.',
    probarKeycloak: 'Test the connection',
    keycloakOk: 'The realm answered and says where its sign-in screen is.',
    numeracion: 'Numbering and how tickets are shared out',
    numeracionAyuda:
      'The prefix of ticket numbers and how new ones are shared out. The installation language arrives with account creation.',
    prefijo: 'Number prefix',
    prefijoAyuda: 'Between 2 and 8 characters, upper case and digits. Example: CS.',
    prefijoAviso:
      'Changing it does not change numbers already issued: tickets that exist keep theirs.',
    repartoPrincipal: 'Main tickets',
    repartoInterno: 'Internal tickets',
    asignacion: 'Sharing out',
    aviso: 'Email notice',
    asignacionNinguna: 'Not shared out',
    asignacionPorTurnos: 'In turn',
    avisoANadie: 'Nobody',
    avisoATodoElEquipo: 'The whole team',
    avisoAlAsignado: 'The assignee',
    guardarNumeracion: 'Save the numbering and the sharing out',
    numeracionGuardada: 'The numbering and the sharing out are saved.',
  },

  menu: {
    misTickets: 'My tickets',
    bandeja: 'Inbox',
    ticketsPrincipales: 'Main tickets',
    ticketsInternos: 'Internal tickets',
    categorias: 'Categories and tags',
    usuarios: 'Users',
    configuracion: 'Settings',
    opciones: 'Options',
    abrir: 'Open the menu',
    cerrar: 'Close the menu',
    plegar: 'Collapse the menu',
    desplegar: 'Expand the menu',
  },

  papeles: {
    usuario: 'User',
    soporte: 'Technical support',
    desarrollo: 'Development',
    administrador: 'Administrator',
  },

  origenes: {
    local: 'Local',
    ad: 'Directory (AD)',
    keycloak: 'Keycloak',
  },

  estados: {
    nuevo: 'New',
    'en progreso': 'In progress',
    'en espera': 'Waiting',
    escalado: 'Escalated',
    resuelto: 'Resolved',
    cerrado: 'Closed',
  },

  estadosDelUsuario: {
    nuevo: 'Received',
    'en progreso': 'In progress',
    'en espera': 'We are waiting for you',
    escalado: 'In progress',
    resuelto: 'Resolved',
    cerrado: 'Closed',
  },

  historial: {
    creado: '{quien} opened the ticket',
    creadoSistema: 'The ticket was opened',
    editado: '{quien} edited the ticket',
    asignado: '{quien} assigned it to {detalle}',
    estado: '{quien} moved it to “{estado}”',
    estadoSistema: 'The ticket went back to “{estado}”',
    escalado: '{quien} escalated it to Development',
    resuelto: '{quien} marked it as resolved',
    cerrado: '{quien} closed it',
    reabierto: '{quien} reopened it',
    observador: '{quien} removed {detalle} from the watchers',
    observador_anadido: '{quien} added {detalle} to the watchers',
    elSistema: 'The system',
  },

  usuarios: {
    titulo: 'Users',
    descripcion: 'The accounts of the installation. This is where they are created and kept.',
    cargando: 'Loading the accounts…',

    // --- the list ---
    nueva: 'New account',
    buscar: 'Search by name or email',
    buscarEtiqueta: 'Search',
    filtroPapel: 'Role',
    filtroOrigen: 'Origin',
    filtroEstado: 'State',
    todos: 'All',
    activas: 'Active',
    desactivadas: 'Deactivated',
    limpiar: 'Clear the filters',
    columnaCuenta: 'Account',
    columnaPapel: 'Role',
    columnaOrigen: 'Origin',
    columnaEstado: 'State',
    columnaUltimaEntrada: 'Last sign-in',
    columnaAcciones: 'Actions',
    activa: 'Active',
    desactivada: 'Deactivated',
    nunca: 'Has never signed in',
    editar: 'Edit',
    nombre: 'First name',
    apellidos: 'Last name',
    abrir: 'Open the record',
    desactivar: 'Deactivate',
    reactivar: 'Reactivate',
    reactivaSola:
      'This is a Keycloak account: it reactivates by itself when that person signs in through their path, and nobody has to do anything here.',
    mandarEnlace: 'Change password',
    sinResultados: 'No account matches those filters.',
    sinCuentas: 'There are no accounts yet. Create the first one.',
    total: 'accounts',
    pagina: 'Page',
    anterior: 'Previous',
    siguiente: 'Next',
    tarjetaVer: 'Open the record',

    // --- creation ---
    altaTitulo: 'New account',
    altaAyuda:
      'The account is created here and, if it is local, the person receives a link to set their password.',
    correo: 'Email',
    papel: 'Role',
    origen: 'Origin',
    idioma: 'Language of the emails',
    origenNota:
      'AD and Keycloak accounts arrive with the directory path: while that does not exist they cannot be created, so that no account is made that nobody could sign in with.',
    idiomaDeLaInstalacion: 'The installation one',
    crear: 'Create the account',
    creando: 'Creating…',
    creada: 'Account created. The link to set the password has been sent.',
    creadaSinEnlace:
      'The account is created, but the invitation link did not go out. Send it again from the list.',
    obligatorios: 'First name, last name and email are required.',

    // --- what the actions answer ---
    mandarEnlacePregunta: 'Send the link to {nombre}?',
    mandarEnlaceAlta: 'They will get an email to set their password. The link expires in 24 hours.',
    mandarEnlaceRecuperacion:
      'They will get an email to set a new password: the one they have now will stop working. The link expires in 1 hour.',
    enlaceMandado: 'Link sent again. It expires in 24 hours.',
    desactivadaHecha: 'Account deactivated. It can no longer sign in.',
    reactivadaHecha: 'Account reactivated. It can sign in again.',
    guardada: 'Changes saved.',

    // --- the record ---
    fichaTitulo: 'The account',
    datos: 'The details',
    estado: 'State',
    ultimaEntrada: 'Last sign-in',
    contrasena: 'Password',
    contrasenaPropia: 'The person sets it from the link they receive.',
    contrasenaDirectorio: 'It is their directory password: it is not changed from here.',
    sinCambios: 'There is nothing to save.',
    volver: 'Back to the list',
    acciones: 'Actions',
    desactivarPregunta: 'Deactivate this account?',
    desactivarConsecuencia:
      'It will not be able to sign in as soon as you do it, even with the browser open: its next request will not be valid. Nothing is deleted, and its tickets still tell their story.',
    siDesactivar: 'Yes, deactivate',
    cambiarOrigen: 'Change the origin',
    cambiarOrigenAyuda:
      'Moving to local removes the link with the directory; if the account is left without a password, the invitation link is sent.',
    guardarOrigen: 'Save the origin',
    origenGuardado: 'Origin changed.',
    guardar: 'Save',
    cancelar: 'Cancel',
  },

  perfil: {
    abrir: 'My profile',
    titulo: 'My profile',
    descripcion: 'Your first name, your last name and the language your emails are written in.',
    nombre: 'First name',
    apellidos: 'Last name',
    correo: 'Email',
    correoAyuda: 'Your email is what you sign in with, and an Administrator changes it.',
    papel: 'Role',
    papelAyuda: 'An Administrator changes your role.',
    idioma: 'Language of the emails',
    idiomaAyuda: 'It is the language of your emails, and the one the application is shown in.',
    guardar: 'Save',
    guardando: 'Saving…',
    guardado: 'Your profile is saved.',
    contrasena: 'Password',
    cambiarContrasena: 'Change my password',
  },

  correos: {
    titulo: 'The emails',
    descripcion:
      'The ten emails the application sends, in both languages. The text is the voice of the institution, and whoever is in charge changes it.',
    queCorreo: 'Which email',
    editado: 'This email has your text, not the factory one.',
    deFabrica: 'This email carries the factory text.',
    idiomaEs: 'Español',
    idiomaEn: 'English',
    asunto: 'Subject',
    cuerpo: 'Body',
    cuerpoAyuda:
      'Write it with the tools above, over the selected text. You can also write the HTML by hand.',
    marcadoresDisponibles: 'Available markers',
    marcadorImprescindible: 'essential',
    faltaMarcador:
      'This email is missing a marker that should be there: {marcadores}. It can be saved like this, but check that it is what you want.',
    vistaPrevia: 'Preview',
    sinVistaPrevia: 'Write something to see how it looks.',
    barraDeFormato: 'Text formatting',
    negrita: 'Bold',
    cursiva: 'Italic',
    lista: 'List',
    enlace: 'Link',
    guardar: 'Save',
    guardando: 'Saving…',
    guardado: 'Saved. The preview already shows the new text.',
    prueba: 'Send me a test',
    pruebaEnviada: 'Test sent to {correo}. Its subject starts with [prueba].',
    restaurar: 'Back to the factory one',
    restaurarPregunta: 'Go back to the factory text?',
    restaurarConsecuencia:
      'The text that is there now in that language is lost and the one the application brings comes back.',
    siRestaurar: 'Yes, go back',
    restaurarHecho: 'This email is back to the factory text.',
    cancelar: 'Cancel',
    cargando: 'Loading the emails…',
  },

  plantillas: {
    'ticket.created': 'New ticket, to the technician',
    'ticket.escalated': 'Escalated ticket, to Development',
    'ticket.waitingUser': 'We need something from the user',
    'ticket.waitingSupport': 'Development needs something',
    'ticket.resolved': 'Ticket resolved, to the user',
    'ticket.closed': 'Ticket closed, to the user',
    'ticket.backToSupport': 'The ticket goes back to Support',
    'ticket.mentioned': 'You were mentioned in a ticket',
    'auth.invitation': 'Account creation',
    'auth.recovery': 'Password recovery',
    'auth.passwordChanged': 'Password changed',
  },

  tickets: {
    nuevo: 'New ticket',
    buscar: 'Search by number, subject or text',
    buscarEtiqueta: 'Search',
    filtroEstado: 'State',
    filtroTipo: 'Type',
    tipoPrincipal: 'Main',
    tipoInterno: 'Internal',
    tipoTodos: 'All',
    todosLosEstados: 'All',
    columnaNumero: 'Ticket',
    columnaAsunto: 'Subject',
    columnaClasificacion: 'Classification',
    columnaEstado: 'State',
    columnaSolicitante: 'Requested by',
    columnaResponsable: 'Assignee',
    columnaActualizado: 'Update date',
    columnaMotivo: 'Reason',
    resumenes: 'What the ticket is about',
    columnaUltimaAccion: 'Last action',

    resumenPendiente: 'Generating…',
    resumenSinMotor: 'No AI engine',
    resumenError: 'Could not summarize',
    resumenVacio: '—',
    editarElAsunto: 'Edit the subject',
    editarLaDescripcion: 'Edit the description',
    anadirEtiquetaAlTicket: 'Add a tag',
    anadirObservador: 'Add a watcher',
    sinPersonasParaObservar: 'There is no active technician or developer to add.',

    etiquetar: 'Mention someone',
    etiquetarAyuda: 'Call another technician or developer: they start watching this ticket.',
    etiquetarA: 'Who are you calling',
    etiquetando: 'Mentioning…',
    sinPersonas: 'There is no active technician or developer to call.',
    observadores: 'Watchers',
    sinObservadores: 'Nobody else is watching this ticket.',
    quitarObservador: 'Remove from the watchers',
    observadorQuitado: '{quien} is no longer watching the ticket.',

    vista: 'View',
    vistaAsignados: 'Assigned to me',
    vistaObservo: 'Watching',
    vistaTodos: 'All',

    // --- the category and the tags (docs/interfaz-y-experiencia.md, section 3.3) ---
    categoria: 'Category',
    etiquetas: 'Tags',
    filtroCategoria: 'Category',
    filtroEtiqueta: 'Tag',
    todasLasCategorias: 'All',
    buscarPorEtiqueta: 'Search by tag',
    sinEtiquetas: 'No tags',

    regenerar: 'Summarize again',
    regenerando: 'Asking for the summaries…',
    resumenesPedidos: 'The two summaries have been requested again.',
    sinResponsable: 'Nobody yet',
    cargando: 'Loading the tickets…',
    vacia: 'No ticket matches those filters.',
    vaciaSinFiltros: 'There are no tickets yet.',
    limpiar: 'Clear the filters',
    total: 'tickets',
    pagina: 'Page',
    anterior: 'Previous',
    siguiente: 'Next',
    verTicket: 'Open',
    verImagen: 'View the image larger',
    verVideo: 'View the video larger',
    tarjetaAsignado: 'Assignee',

    nuevoTitulo: 'New ticket',
    nuevoAyuda:
      'Tell us what is happening, with as much detail as you can. Attach a screenshot if you have one.',
    asunto: 'Subject',
    campoDescripcion: 'Description',
    camposDescripcionAyuda:
      'Write with formatting and put the files where you want them to show: they are shown right here.',
    archivoRechazado: 'These files were not added:',
    paraQuien: 'Who is the ticket for?',
    paraQuienAyuda: 'Support can create a ticket on behalf of someone else.',
    elegirSolicitante: 'The person',
    crear: 'Create the ticket',
    creando: 'Creating…',
    creado: 'Ticket created: {numero}.',
    creadoConAdjuntos: 'Ticket created: {numero}. These attachments did not upload:',
    obligatorios: 'The subject and the description are required.',
    volver: 'Back',
    volverALaBandeja: 'Back to the inbox',

    // The new-ticket screen introduces the classification: **the category is required** —there is no
    // ticket without one— and the tags are optional and normalise themselves as they are typed.
    elegirCategoria: 'Choose a category',
    categoriaAyuda: 'Every ticket is classified: Support and Development group by this.',
    sinCategoriasActivas:
      'There is no active category, so a ticket cannot be created right now. Ask Support or an Administrator to create one.',
    etiquetasCampo: 'Tags',
    guardandoEtiquetas: 'Saving…',
    sinEtiquetasQueCoincidan: 'No tag matches that search.',
    editarEtiquetas: 'Edit the tags',
    etiquetasAyuda:
      'Optional. Lowercase with hyphens (red-wifi), up to 32 characters; press Enter or comma to add them.',
    etiquetasSugeridas: 'Already exist',
    anadirEtiqueta: 'Add the tag',
    quitarEtiqueta: 'Remove the tag',

    copiarNumero: 'Copy the number',
    copiado: 'Copied',
    copiarNoSePudo: 'It could not be copied for you: select the number and copy it yourself.',
    conversacion: 'The conversation',
    ficha: 'The details',
    solicitante: 'Requested by',
    creadoPor: 'Opened by',
    responsable: 'Assignee',
    abiertoEl: 'Opened on',
    actualizadoEl: 'Last update',
    resueltoEl: 'Resolved on',
    cerradoEl: 'Closed on',

    // The details panel shows the classification and, for whoever can edit the ticket, lets them
    // change it with the same pattern as the assignee (docs/interfaz-y-experiencia.md, section 3.4).
    cambiarCategoria: 'Change the category',
    guardarClasificacion: 'Save the category and the tags',
    clasificacionGuardada: 'Category and tags saved.',
    sinEtiquetasFicha: 'No tags.',
    motivoDelEscalado: 'Reason for the escalation',
    elPrincipal: 'The main ticket',
    sinAdjuntos: 'No attachments',
    escribirComentario: 'Write a comment',
    // A comment can be just an attachment: it is said, because otherwise it looks as if writing were
    // required. And the three doors for attaching are told here, which is where the writing happens.
    comentarioAyuda:
      'With formatting, and with the files inside: drag them here, paste them with Ctrl+V or look for them. A comment can carry only files.',
    enviar: 'Comment',
    enviando: 'Sending…',
    comentarioVacio: 'Write something or attach a file before sending it.',
    comentarioEditado: 'edited',
    comentarioBorrado: 'This comment was deleted. Its text is gone.',
    editar: 'Edit',
    borrar: 'Delete',
    guardar: 'Save',
    cancelar: 'Cancel',
    borrarPregunta: 'Delete this comment?',
    borrarConsecuencia: 'The text is wiped for good and the mark stays. The attachments stay too.',
    siBorrar: 'Yes, delete',
    cargandoTicket: 'Loading the ticket…',
    sinComentarios: 'There are no comments yet.',

    vistaPrincipal: 'Main',
    vistaDoble: 'Both',
    vistaInterna: 'Internal',
    vistaEtiqueta: 'What you want to see',

    asignarA: 'Assign to',
    asignar: 'Assign',
    asignado: 'Ticket assigned.',
    empezar: 'Start',
    preguntarAlUsuario: 'Ask the user',
    resolver: 'Resolve',
    resolverTitulo: 'Resolve the ticket',
    resolverAyuda:
      'Tell the user what was done, in their own language: it is the only thing they will read.',
    resolviendo: 'Resolving…',
    cerrar: 'Close',
    cerrarPregunta: 'Close this ticket?',
    estadoDelTicket: 'Ticket status',
    preguntarConsecuencia: 'The ticket waits for the user to answer, and they are told.',
    empezarConsecuencia: 'The ticket moves to in progress: whoever works it has it.',
    pasarA: 'Move the ticket to "{estado}"?',
    confirmar: 'Confirm',
    comentarioDelCierre: 'Closing comment',
    cierreAyuda: 'It stays in the conversation: it is what is read to know why it was closed.',
    cerrarConsecuencia: 'It will stop taking comments. It can be reopened whenever it is needed.',
    siCerrar: 'Yes, close',
    reabrir: 'Reopen',
    reabrirAyuda: 'This ticket is closed: to carry on talking it has to be reopened.',
    accionesDelTicket: 'Ticket actions',
    queHacer: 'What do you want to do?',
    escalar: 'Escalate',
    escalarTitulo: 'Escalate',
    escalarAyuda:
      'Say what you tried, what you ruled out and what is needed. Without the reason, this is just passing it on.',
    motivo: 'Reason',
    escalarBoton: 'Escalate',
    laPregunta: 'The question',
    enviarLaPregunta: 'Send the question',
    loQueNecesitas: 'What you need from Support',
    pedirlo: 'Ask for it',
    queSeHizo: 'What was done',
    resolverYAvisar: 'Resolve and tell them',
    porQueNoEsCodigo: 'Why this is not a code change',
    escalando: 'Escalating…',
    pedirAlgo: 'I need something from Support',
    devolver: 'This is not a code change',
    devolverPregunta: 'Send the case back to Support?',
    devolverConsecuencia:
      'The main ticket goes back to the Support inbox, which is who talks to the user, and they are told.',
    siDevolver: 'Yes, send it back',
    sinAcciones: 'There is nothing to do here: this ticket is read-only.',

    ticketActualizado: 'Ticket updated.',
    categoriaCambiada: 'Category changed.',
    etiquetasCambiadas: 'Tags saved.',
    observadorAnadido: 'Watcher added.',
    descripcionActualizada: 'Description updated.',
    estadoCambiado: 'The ticket has changed state.',
    resueltoAviso: 'Ticket resolved. The user has been told.',
    cerradoAviso: 'Ticket closed.',
    reabiertoAviso: 'Ticket reopened.',
    escaladoAviso: 'Ticket escalated. Development has been told.',
    comentarioEscrito: 'Comment written.',
    // The comment is already published and a file did not make it: it says which, and it can be
    // retried without writing the comment again, which is what stops the duplicate comment.
    adjuntosQueFaltan: 'The comment was published, but these files could not be uploaded:',
    reintentarAdjuntos: 'Retry the upload',
    adjuntosSubidos: 'Every file of the comment is now uploaded.',
    comentarioGuardado: 'Comment saved.',
    comentarioBorradoAviso: 'Comment deleted.',
    adjuntoSubido: 'Attachment uploaded.',
    archivoNoTraido: 'The file could not be fetched.',
    // The gap left by a file that was named in the text and never made it (decision 44): it is shown
    // where it was, and not as a silent hole that looks like a broken screen.
    adjuntoNoSubio: 'Did not upload',
    descargar: 'Download',
    abrirEnPestana: 'Open in a tab',
    ver: 'View',
    vistaPrevia: 'Preview',
    cerrarVistaPrevia: 'Close the preview',
    // What is shown while an attachment is being fetched to be painted inside the text.
    trayendoAdjunto: 'Fetching the file…',

    // --- "Categories and tags": the catalogue (docs/interfaz-y-experiencia.md, section 3.7) ---
    buscarTitulo: 'Search tickets',
    abrirBusqueda: 'Search',
    buscarAyuda: 'Every criterion can be combined: whatever comes back meets all of them.',
    resumenFiltros: '{n} filters',
    resumenUnFiltro: '1 filter',
    sinFiltros: 'No filters',
    quitarLosFiltros: 'Clear the filters',

    nuevaEtiqueta: 'New tag',
    creandoEtiqueta: 'Creating…',
    crearEtiqueta: 'Create',
    renombrarEtiqueta: 'Rename',
    retirarEtiqueta: 'Retire',
    etiquetaCreada: 'Tag created.',
    etiquetaRenombrada: 'Tag renamed on every ticket that carries it.',
    etiquetaRetirada: 'Tag retired and removed from its tickets.',
    avisoDeRetiradaDeEtiqueta:
      '“{etiqueta}” will be retired and removed from {n} tickets. It can be written again on any ticket.',
    etiquetaSinTickets: 'No ticket carries it yet.',
    nombreDeEtiqueta: 'Tag',
    etiquetasDescripcion:
      'This is what tells apart what a category does not. Tags are born when someone writes them on a ticket, and here they are fixed and retired.',
    categoriasTitulo: 'Categories and tags',
    categoriasDescripcion:
      'Tickets are classified with this. Retiring a category stops offering it when creating, but does not touch the tickets that already have it.',
    categoriasLista: 'Categories',
    nuevaCategoria: 'New category',
    nombreDeCategoria: 'Name',
    crearCategoria: 'Create',
    creandoCategoria: 'Creating…',
    renombrarCategoria: 'Rename',
    retirarCategoria: 'Retire',
    reactivarCategoria: 'Put back',
    categoriaActiva: 'Active',
    categoriaRetirada: 'Retired',
    sinCategorias: 'There is no category yet.',
    sinEtiquetasEnUso: 'There is no tag in use yet.',
    categoriaCreada: 'Category created.',
    categoriaRenombrada: 'Category renamed.',
    categoriaRetiradaAviso: 'Category retired. The tickets that have it keep it.',
    categoriaReactivadaAviso: 'Category put back.',
  },

  // The buttons of the rich text editor, which is the box where a ticket and a comment are written
  // (docs/interfaz-y-experiencia.md, section 6.5).
  editor: {
    barra: 'Text formatting',
    negrita: 'Bold',
    cursiva: 'Italic',
    subrayado: 'Underline',
    tachado: 'Strikethrough',
    listaVinietas: 'Bulleted list',
    listaNumerada: 'Numbered list',
    enlace: 'Link',
    adjuntar: 'Attach',
    direccion: 'The address of the link',
    direccionPoner: 'Set the link',
    direccionCancelar: 'Cancel',
    // The link is checked before it is set: whatever does not start with one of the three schemes is
    // refused, and it is the same rule the backend applies when saving (docs/modules/tickets.md, 2.3).
    direccionInvalida: 'The address has to start with http, https or mailto.',
  },

  tema: {
    etiqueta: 'Theme',
    claro: 'Light',
    oscuro: 'Dark',
    papel: 'Paper',
    niebla: 'Mist',
    contraste: 'High contrast',
    grafito: 'Graphite',
    noche: 'Night',
    sepia: 'Sepia',
    claros: 'Light',
    oscuros: 'Dark',
  },

  entrada: {
    titulo: 'Sign in',
    descripcion: 'Catalina Support',
    correo: 'Email address',
    correoAyuda: 'Use the address your account was created with.',
    contrasena: 'Password',
    entrar: 'Sign in',
    entrando: 'Signing in…',
    olvidada: 'I have forgotten my password',
    obligatorios: 'Enter your email address and your password.',
    // El camino de Keycloak: el botón se enseña **sólo si esta instalación lo tiene**, y lo dice el
    // backend (docs/modules/auth.md, decisión 27).
    oCon: 'Or sign in with your organization account',
    conKeycloak: 'Sign in with Keycloak',
    // With the method in AD the form **is the same** —email and password—, but the password is not
    // ours: it is said, and the recovery link goes away, since it would be of no use.
    porDirectorio:
      'This installation signs in with the organization accounts: enter your usual email address and password.',
    soloKeycloak:
      'This installation signs in with Keycloak: you sign in from its screen, not from here.',
    // The factory account door: with the method in Keycloak there is no form, and **without it the
    // installation would be left with nobody able to change the method back**.
    comoAdministrador: 'Sign in as administrator',
    cuentaDeFabrica: 'The factory account password lives in the installation settings.',
  },

  olvido: {
    titulo: 'Reset your password',
    explicacion:
      'Enter your email address and we will send you a link to set a new password. The link expires in 1 hour.',
    correo: 'Email address',
    enviar: 'Send the link',
    enviando: 'Sending…',
    enviado:
      'If that address has an account, a link is on its way. Check your email, and if it is not there, look in the spam folder.',
    volver: 'Back to the sign-in screen',
  },

  establecer: {
    titulo: 'Set your password',
    explicacion: 'Type the password you want to use, twice.',
    nueva: 'New password',
    repetir: 'Repeat the password',
    guardar: 'Save the password',
    guardando: 'Saving…',
    ayuda: 'At least 8 characters.',
    sinToken:
      'This link does not carry what it needs. Ask for another one from the sign-in screen.',
    noCoinciden: 'The two passwords are not the same.',
    hecha: 'Your password is set. Sign in with it.',
  },

  cambio: {
    titulo: 'Change my password',
    actual: 'Current password',
    nueva: 'New password',
    repetir: 'Repeat the new password',
    guardar: 'Change the password',
    guardando: 'Changing…',
    ayuda: 'At least 8 characters.',
    noCoinciden: 'The two new passwords are not the same.',
    hecha: 'Your password has changed. We have sent you a notice by email.',
  },

  sinPermiso: {
    titulo: 'You cannot see this screen',
    explicacion:
      'Your account is not allowed to do what you asked for. If you think this is a mistake, talk to Support.',
    volver: 'Back to my work',
  },

  servidor: {
    titulo: 'No connection to the server',
    explicacion:
      'We could not reach the {instalacion} server. It may be a momentary glitch: try again in a few seconds.',
    reintentar: 'Try again',
  },

  sesion: {
    salir: 'Sign out',
    salirPregunta: 'Sign out of the application?',
    salirConsecuencia: 'You will have to sign in again with your account.',
    caducada: 'Your session has expired. Sign in again.',
  },

  comun: {
    volver: 'Back',
    cancelar: 'Cancel',
    cargando: 'Loading…',
    obligatorio: 'This field is required.',
  },

  errores: {
    'error interno': 'Something failed on our side. Please try again in a moment.',

    'ai.unavailable': 'The summary engine is not available. Tickets work just the same.',
    'ai.invalid': 'The summary engine answered something we cannot read.',
    'tickets.mention.notAllowed': 'You cannot mention anyone in this ticket.',

    'auth.invalidCredentials': 'The email address or the password is not correct.',
    'auth.accountInactive': 'Your account is deactivated. Talk to Support to get it activated.',
    'auth.session.invalid': 'Your session is not valid. Sign in again.',
    'auth.session.expired': 'Your session has expired. Sign in again.',
    'auth.forbidden': 'Your account is not allowed to do that.',
    'auth.token.expired': 'That link has expired. Ask for a new one.',
    'auth.token.used':
      'That link no longer works: it may have been used already, or replaced by a newer one.',
    'auth.password.tooShort': 'The password must be at least 8 characters long.',
    'auth.password.wrong': 'The current password is not the one you typed.',
    'auth.password.notLocal':
      'That password is not changed from here: it belongs to your work directory.',
    'auth.oidc.state': 'The sign-in response is not valid or has expired. Please try again.',
    'auth.oidc.notConfigured': 'This installation has no Keycloak configured.',
    'auth.oidc.rejected': 'The sign-in with Keycloak could not be completed.',
    'auth.oidc.unavailable':
      'Keycloak is not responding. Try again in a moment or sign in with your e-mail and password.',
    'auth.oidc.noEmail':
      'Your organization account has no e-mail address, and without one the account cannot be created here. Tell Support.',
    'auth.directory.unavailable':
      'Your organization directory is not responding. Try again in a moment.',
    'auth.directory.rejected': 'The directory rejected those credentials.',

    'settings.name.tooLong': 'The name cannot be longer than 60 characters.',
    'settings.language.unknown': 'That language does not exist.',
    'settings.primaryColor.invalid':
      'That colour does not work: it has to be a hexadecimal colour, like #1d4ed8.',
    'settings.numberPrefix.invalid':
      'The prefix must be 2 to 8 characters, uppercase letters and digits.',
    'settings.assignment.unknown': 'That way of handing out tickets does not exist.',
    'settings.notification.unknown': 'That way of notifying does not exist.',
    'settings.notification.withoutAssignment':
      'You cannot notify only the assignee if the ticket is not handed out: there would be nobody to notify.',
    'settings.logo.format': 'That file is not one of the accepted images: PNG, JPEG, WebP or SVG.',
    'settings.logo.invalid': 'That image could not be read, or it is bigger than allowed.',
    'settings.logo.tooBig': 'That logo is too heavy: the maximum is 1 MB.',
    'settings.logo.variantUnknown': 'That logo slot does not exist: it is «light» or «dark».',
    'settings.logo.notFound':
      'The installation has no logo of its own: the factory one is being used.',
    'settings.notFound': 'The installation settings could not be found.',
    'settings.method.unknown':
      'That sign-in method does not exist: it is “local”, “ad” or “keycloak”.',
    'settings.method.notConfigured':
      'That method is not configured yet, and choosing it would leave the installation with no door for everybody except the factory account. Configure it first.',
    'settings.directory.incomplete': 'The directory is missing its server or its search base.',
    'settings.keycloak.incomplete':
      'Keycloak is missing its issuer, its client or its redirect address.',
    'settings.directory.unreachable':
      'The directory could not be reached: check the server, the port and the service account.',
    'settings.keycloak.unreachable':
      'The Keycloak realm could not be read: check the issuer and that the realm exists.',

    'users.notFound': 'That account does not exist.',
    'users.email.duplicated': 'That email address is already used by another account.',
    'users.email.invalid': 'That does not look like an email address.',
    'users.name.required': 'The first or the last name is missing.',
    'users.role.notAllowed': 'You are not allowed to hand out that role.',
    'users.origin.unknown': 'That account origin does not exist.',
    'users.language.unknown': 'That language does not exist.',
    'users.directory.notFound':
      'The directory does not know that person: they are no longer there.',
    'users.origin.byDirectory':
      'Directory accounts are neither created nor moved to that origin by hand: whoever is in the directory signs in with their account and theirs is brought up to date on that first sign-in.',
    'users.directory.activatesItself':
      'That account signs in by itself through Keycloak: as soon as that person signs in, it will be active again and nobody has to do anything.',
    'users.directory.unavailable':
      'The directory could not be reached to check. Try again in a moment.',
    'users.origin.inUse': 'The origin of an active directory account cannot be changed.',
    'users.password.notLocal': 'That account\'s password is checked by the directory: it is not changed here.',
    'settings.timeZone.unknown': 'That time zone does not exist. Pick one from the list.',
    'settings.publicUrl.invalid':
      'The address must start with http:// or https:// and carry its server.',
    'auth.publicUrl.missing':
      'There is no public address configured: without it the link for the e-mail cannot be built. It is configured in Settings.',
    'users.accountInactive': 'That account is deactivated: activate it before sending it a link.',
    'users.selfDeactivation': 'You cannot deactivate your own account: ask someone else to do it.',
    'users.directoryMayReturn':
      'This is a directory account: if it is still active there, it will sign in again and reactivate itself.',
    'users.stateHasItsOwnAction':
      'An account is activated or deactivated with its own action, not by editing the account.',

    'tickets.notFound': 'That ticket does not exist.',
    'tickets.subject.required': 'Write a subject.',
    'tickets.description.required': 'Write a description.',
    // The backend sanitiser only accepts the tags and attributes of its whitelist, and whatever is not
    // there **is not saved**: it is said, because a text saved halfway is worse than one that is
    // refused (docs/modules/tickets.md, section 2.3). With the editor pasting without formatting, what
    // reaches here has the formatting of whatever was pasted.
    'tickets.body.notAllowed':
      'The text carries something that is not allowed. Remove the formatting from what you pasted and try again.',
    'tickets.reason.required':
      'Write why it is escalated: without the reason, Development does not know what is needed.',
    'tickets.comment.required': 'Write a comment.',
    'tickets.state.unknown': 'That state does not exist.',
    'tickets.transition.notAllowed': 'That state change cannot be done that way.',
    'tickets.forbidden': 'That is not yours to do on this ticket.',
    'tickets.closed': 'The ticket is closed: reopen it to carry on.',
    'tickets.comment.notYours': 'Only whoever wrote a comment can change it.',
    'tickets.comment.notFound': 'That comment does not exist.',
    'tickets.assignee.unknown': 'That person cannot be the assignee of this ticket.',
    'tickets.requester.unknown': 'That person does not exist or is deactivated.',
    'tickets.requester.notAllowed': 'Only Support can create a ticket on behalf of someone else.',
    'tickets.attachment.extension': 'That kind of file is not allowed.',
    'tickets.attachment.tooBig': 'That file is too heavy: the maximum is 25 MB.',
    'tickets.attachment.invalid': 'The file could not be read.',
    'tickets.attachment.notFound': 'That attachment does not exist.',

    // The category and tag keys (docs/modules/tickets.md, section 2.3.2).
    'tickets.category.required': 'A category is required: there is no ticket without one.',
    'tickets.category.notFound': 'That category does not exist.',
    'tickets.category.inactive':
      'That category is retired: choose another one to classify the ticket.',
    'tickets.category.name.required': 'Write a name for the category.',
    'tickets.category.duplicate': 'There is already a category with that name.',
    'tickets.category.lastActive':
      'The last active category cannot be retired: the catalogue cannot be left with none.',
    'tickets.tag.forbidden': 'Only the Administrator can retire a tag.',
    'tickets.tag.duplicate': 'That tag already exists.',
    'tickets.cierre.sinComentario': 'To close the ticket you have to say why.',
    'tickets.etiqueta.desconocida': 'That tag is not in the catalogue. Only the Administrator can create new tags.',
    'tickets.tag.notFound': 'That tag does not exist.',
    'tickets.category.forbidden': 'Only an Administrator can retire or put back a category.',
    'tickets.tag.required': 'Write something for the tag: it cannot be empty.',
    // The field itself stops it, but if something arrives from another door it is said all the same.
    'tickets.tag.tooLong': 'A tag cannot be longer than 32 characters.',

    'mail.template.notFound': 'That template does not exist.',
    'mail.template.unknownKey': 'That is not one of the ten emails.',
    'mail.language.unknown': 'That language does not exist.',
    'mail.subject.required': 'The subject cannot be empty.',
    'mail.body.required': 'The body cannot be empty.',
    'mail.marker.unknown': 'The text uses a placeholder that email does not accept.',
    'mail.marker.missing': 'A placeholder used by the text has no value.',
    'mail.test.noEmail':
      'The factory account has no email address, so there is nowhere to send the test.',
    'mail.smtp.notConfigured': 'No mail server is configured.',
    'mail.smtp.connect': 'Could not connect to the mail server.',
    'mail.smtp.handshake': 'The mail server did not accept the encrypted connection.',
    'mail.smtp.auth': 'The mail server rejected the credentials.',
    'mail.smtp.from': 'The mail server rejected the sender.',
    'mail.smtp.recipient': 'The mail server rejected the recipient.',
    'mail.smtp.data': 'The mail server cut the message off.',
    'mail.smtp.write': 'Could not write the message to the mail server.',
    'mail.smtp.close': 'The mail server closed the connection before finishing.',
    'mail.to.empty': 'There are no recipients.',
    'mail.to.invalid': 'One of the recipient addresses is not valid.',
    'mail.from.invalid': 'The sender address is not valid.',
  },
};
