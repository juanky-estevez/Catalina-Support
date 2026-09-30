import { bootstrapApplication } from '@angular/platform-browser';

import { appConfig } from './app/app.config';
import { App } from './app/app';
import { BrandService } from './app/core/services/brand.service';
import { aplicarTemaInicial } from './app/core/services/theme.service';

// El tema, antes de arrancar: quien haya elegido uno distinto del de su sistema no ve el parpadeo
// del tema equivocado al cargar. Se hace aquí y no con un script en el `index.html` porque la CSP
// de producción es `script-src 'self'` (docs/interfaz-y-experiencia.md, sección 6.2).
aplicarTemaInicial();

bootstrapApplication(App, appConfig)
  .then((aplicacion) => {
    // La marca de la instalación (el logo y el color institucional) se pide **en cuanto arranca**,
    // sin esperarla: si tarda o falla, se enseñan los de fábrica y la aplicación funciona igual
    // (docs/modules/settings.md, sección 5.4).
    void aplicacion.injector.get(BrandService).cargar();
  })
  .catch((err) => console.error(err));
