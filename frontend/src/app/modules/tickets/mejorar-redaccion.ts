import { Component, inject, input, output, signal } from '@angular/core';

import { TranslationService } from '../../core/i18n/translation.service';
import { Aviso } from '../../shared/components/aviso';
import { Boton } from '../../shared/components/boton';
import { Dialogo } from '../../shared/components/dialogo';
import { Selector, type OpcionSelector } from '../../shared/components/selector';
import { claveDelError } from '../../shared/errores';
import {
  TicketsService,
  type EditorDeRedaccion,
  type TonoDeRedaccion,
} from './tickets.service';

@Component({
  selector: 'app-mejorar-redaccion',
  imports: [Aviso, Boton, Dialogo, Selector],
  template: `
    <app-dialogo
      [titulo]="t().tickets.iaRedaccionTitulo"
      [etiquetaCerrar]="t().tickets.cancelar"
      [abierto]="abierto()"
      (cerrado)="cerrar()"
    >
      <div class="flex flex-col gap-4">
        <p class="text-sm text-apagado">{{ t().tickets.iaRedaccionAyuda }}</p>
        <p class="text-sm text-apagado">{{ t().tickets.iaRedaccionPrivacidad }}</p>

        <div class="flex flex-col gap-1">
          <label for="ia-redaccion-borrador" class="text-sm font-medium text-texto">
            {{ t().tickets.iaRedaccionBorrador }}
          </label>
          <textarea
            id="ia-redaccion-borrador"
            rows="10"
            class="w-full rounded-md border border-borde-fuerte bg-superficie px-3 py-2 text-sm text-texto"
            [value]="borrador()"
            [disabled]="generando()"
            (input)="cambiarBorrador($event)"
          ></textarea>
        </div>

        <app-selector
          identificador="ia-redaccion-tono"
          [etiqueta]="t().tickets.iaRedaccionTono"
          [opciones]="tonos()"
          [deshabilitado]="generando()"
          [(valor)]="tono"
        />

        @if (error()) {
          <app-aviso forma="error" [texto]="error()" />
        }
        @if (resultadoListo()) {
          <p class="text-sm text-exito" role="status">{{ t().tickets.iaRedaccionLista }}</p>
        } @else if (generando()) {
          <p class="text-sm text-apagado" role="status">{{ t().tickets.iaRedaccionGenerando }}</p>
        }

        <div class="flex flex-wrap gap-2">
          @if (resultadoListo()) {
            <app-boton [texto]="t().tickets.iaRedaccionUsar" (pulsado)="usar()" />
          }
          <app-boton
            [forma]="resultadoListo() ? 'secundario' : 'principal'"
            [texto]="resultadoListo() ? t().tickets.iaRedaccionRegenerar : t().tickets.iaRedaccionMejorar"
            [textoTrabajando]="t().tickets.iaRedaccionGenerando"
            [trabajando]="generando()"
            [deshabilitado]="!borrador().trim()"
            (pulsado)="generar()"
          />
          <app-boton forma="secundario" [texto]="t().tickets.cancelar" (pulsado)="cerrar()" />
        </div>
      </div>
    </app-dialogo>
  `,
})
export class MejorarRedaccion {
  readonly numero = input.required<string>();
  readonly aplicado = output<string>();

  private readonly tickets = inject(TicketsService);
  private readonly textos = inject(TranslationService);
  protected readonly t = this.textos.textos;

  protected readonly abierto = signal(false);
  protected readonly borrador = signal('');
  protected readonly tono = signal<TonoDeRedaccion>('professional');
  protected readonly generando = signal(false);
  protected readonly resultadoListo = signal(false);
  protected readonly error = signal('');
  private editor: EditorDeRedaccion = 'comment';
  private generacion = 0;

  protected readonly tonos = signal<readonly OpcionSelector[]>([]);

  constructor() {
    this.tonos.set(this.opcionesDeTono());
  }

  abrirCon(texto: string, editor: EditorDeRedaccion): void {
    if (!texto.trim()) return;
    this.editor = editor;
    this.borrador.set(texto);
    this.tono.set('professional');
    this.resultadoListo.set(false);
    this.error.set('');
    this.tonos.set(this.opcionesDeTono());
    this.abierto.set(true);
  }

  protected cambiarBorrador(evento: Event): void {
    this.borrador.set((evento.target as HTMLTextAreaElement).value);
    this.resultadoListo.set(false);
  }

  protected async generar(): Promise<void> {
    const texto = this.borrador().trim();
    if (!texto || this.generando()) return;
    const turno = ++this.generacion;
    this.generando.set(true);
    this.error.set('');
    try {
      const respuesta = await this.tickets.mejorarRedaccion(
        this.numero(), this.editor, texto, this.tono(),
      );
      if (turno !== this.generacion || !this.abierto()) return;
      this.borrador.set(respuesta.text);
      this.resultadoListo.set(true);
    } catch (error) {
      if (turno !== this.generacion || !this.abierto()) return;
      this.error.set(this.textos.error(claveDelError(error)));
    } finally {
      if (turno === this.generacion) this.generando.set(false);
    }
  }

  protected usar(): void {
    if (!this.resultadoListo()) return;
    this.aplicado.emit(this.borrador());
    this.cerrar();
  }

  protected cerrar(): void {
    this.generacion++;
    this.generando.set(false);
    this.abierto.set(false);
  }

  private opcionesDeTono(): readonly OpcionSelector[] {
    const t = this.t().tickets;
    return [
      { valor: 'professional', etiqueta: t.iaTonoProfesional, grupo: t.iaRedaccionTono },
      { valor: 'friendly', etiqueta: t.iaTonoCordial, grupo: t.iaRedaccionTono },
      { valor: 'brief', etiqueta: t.iaTonoBreve, grupo: t.iaRedaccionTono },
      { valor: 'empathetic', etiqueta: t.iaTonoEmpatico, grupo: t.iaRedaccionTono },
      { valor: 'technical', etiqueta: t.iaTonoTecnico, grupo: t.iaRedaccionTono },
    ];
  }
}
