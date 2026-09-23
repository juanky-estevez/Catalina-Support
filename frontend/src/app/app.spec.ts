import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { App } from './app';

describe('App', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [App],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();
  });

  it('se crea', () => {
    const fixture = TestBed.createComponent(App);
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('muestra el título y el estado del servicio', async () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();

    // El armazón comprueba la salud del backend al arrancar: hay que responder a esa
    // petición para que la prueba no quede esperando.
    TestBed.inject(HttpTestingController)
      .expectOne('/api/health')
      .flush({ status: 'ok', database: 'ok' });

    await fixture.whenStable();

    const compilado = fixture.nativeElement as HTMLElement;
    expect(compilado.querySelector('h1')?.textContent).toContain('Catalina Support');
    expect(compilado.textContent).toContain('servicio disponible');
  });
});
