import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { type ComponentFixture, TestBed } from '@angular/core/testing';

import { TranslationService } from '../../core/i18n/translation.service';
import { MejorarRedaccion } from './mejorar-redaccion';

beforeAll(() => {
  HTMLDialogElement.prototype.showModal = function (this: HTMLDialogElement) {
    this.open = true;
  };
  HTMLDialogElement.prototype.close = function (this: HTMLDialogElement) {
    this.open = false;
    this.dispatchEvent(new Event('close'));
  };
});

describe('MejorarRedaccion', () => {
  let fixture: ComponentFixture<MejorarRedaccion>;
  let http: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MejorarRedaccion],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    TestBed.inject(TranslationService).cambiar('es');
    http = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(MejorarRedaccion);
    fixture.componentRef.setInput('numero', 'CS-2026-0001');
    fixture.detectChanges();
  });

  afterEach(() => http.verify());

  it('permite revisar el resultado y sólo lo aplica cuando se pulsa Usar este texto', async () => {
    let aplicado = '';
    fixture.componentInstance.aplicado.subscribe((texto) => (aplicado = texto));
    fixture.componentInstance.abrirCon('hola necesito datos', 'comment');
    fixture.detectChanges();

    const host = fixture.nativeElement as HTMLElement;
    expect(host.querySelector<HTMLDialogElement>('dialog')?.open).toBe(true);
    expect(host.querySelector<HTMLTextAreaElement>('#ia-redaccion-borrador')?.value).toBe(
      'hola necesito datos',
    );

    const mejorar = [...host.querySelectorAll<HTMLButtonElement>('button')].find(
      (boton) => boton.textContent?.trim() === 'Mejorar',
    );
    mejorar!.click();

    const peticion = http.expectOne('/api/tickets/CS-2026-0001/writing/improve');
    expect(peticion.request.method).toBe('POST');
    expect(peticion.request.body).toEqual({
      editor: 'comment',
      draft: 'hola necesito datos',
      tone: 'professional',
    });
    peticion.flush({ text: 'Hola, ¿podría compartir los datos necesarios?' });

    await fixture.whenStable();
    fixture.detectChanges();
    expect(aplicado).toBe('');
    expect(host.querySelector<HTMLTextAreaElement>('#ia-redaccion-borrador')?.value).toBe(
      'Hola, ¿podría compartir los datos necesarios?',
    );

    const usar = [...host.querySelectorAll<HTMLButtonElement>('button')].find((boton) =>
      boton.textContent?.includes('Usar este texto'),
    );
    usar!.click();
    expect(aplicado).toBe('Hola, ¿podría compartir los datos necesarios?');
  });

  it('mantiene el borrador y muestra el error dentro del modal cuando falla la IA', async () => {
    fixture.componentInstance.abrirCon('Texto que debe conservarse', 'description');
    fixture.detectChanges();

    const host = fixture.nativeElement as HTMLElement;
    const mejorar = [...host.querySelectorAll<HTMLButtonElement>('button')].find(
      (boton) => boton.textContent?.trim() === 'Mejorar',
    );
    mejorar!.click();
    http.expectOne('/api/tickets/CS-2026-0001/writing/improve').flush(
      { error: 'tickets.writing.unavailable' },
      { status: 503, statusText: 'Unavailable' },
    );

    await fixture.whenStable();
    fixture.detectChanges();
    expect(host.querySelector<HTMLTextAreaElement>('#ia-redaccion-borrador')?.value).toBe(
      'Texto que debe conservarse',
    );
    expect(host.textContent).toContain('El motor de IA no está disponible. El borrador se conserva.');
  });
});
