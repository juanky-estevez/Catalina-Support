import { TestBed } from '@angular/core/testing';

import { Toast } from './toast';

describe('Toast', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [Toast] }).compileComponents();
  });

  afterEach(() => vi.useRealTimers());

  function montar(forma: 'informacion' | 'exito' | 'error' | 'atencion') {
    const fixture = TestBed.createComponent(Toast);
    fixture.componentRef.setInput('mensaje', { forma, texto: 'Resultado de la acción' });
    fixture.componentRef.setInput('etiquetaCerrar', 'Cerrar');
    fixture.detectChanges();
    return fixture;
  }

  it('anuncia el resultado, permite cerrarlo y permanece fijo en la ventana', () => {
    const fixture = montar('error');
    const cerrado = vi.fn();
    fixture.componentInstance.cerrado.subscribe(cerrado);

    const host = fixture.nativeElement as HTMLElement;
    expect(host.classList.contains('fixed')).toBe(true);
    expect(host.querySelector('[role="alert"]')?.textContent).toContain('Resultado de la acción');

    const boton = host.querySelector('button[aria-label="Cerrar"]') as HTMLButtonElement;
    boton.click();
    expect(cerrado).toHaveBeenCalledOnce();
  });

  it('retira éxito e información a los cinco segundos', async () => {
    vi.useFakeTimers();
    for (const forma of ['exito', 'informacion'] as const) {
      const fixture = montar(forma);
      const cerrado = vi.fn();
      fixture.componentInstance.cerrado.subscribe(cerrado);

      await vi.advanceTimersByTimeAsync(4_999);
      expect(cerrado).not.toHaveBeenCalled();
      await vi.advanceTimersByTimeAsync(1);
      expect(cerrado).toHaveBeenCalledOnce();
      fixture.destroy();
    }
  });

  it('mantiene error y atención hasta el cierre manual', async () => {
    vi.useFakeTimers();
    for (const forma of ['error', 'atencion'] as const) {
      const fixture = montar(forma);
      const cerrado = vi.fn();
      fixture.componentInstance.cerrado.subscribe(cerrado);

      await vi.advanceTimersByTimeAsync(30_000);
      expect(cerrado).not.toHaveBeenCalled();
      expect((fixture.nativeElement as HTMLElement).querySelector('[role="alert"]')).not.toBeNull();
      fixture.destroy();
    }
  });

  it('reinicia el tiempo cuando un resultado sustituye al anterior', async () => {
    vi.useFakeTimers();
    const fixture = montar('exito');
    const cerrado = vi.fn();
    fixture.componentInstance.cerrado.subscribe(cerrado);

    await vi.advanceTimersByTimeAsync(4_000);
    fixture.componentRef.setInput('mensaje', { forma: 'exito', texto: 'Resultado nuevo' });
    fixture.detectChanges();
    await vi.advanceTimersByTimeAsync(4_000);
    expect(cerrado).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(1_000);
    expect(cerrado).toHaveBeenCalledOnce();
  });
});
