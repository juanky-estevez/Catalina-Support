import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { routes } from './app.routes';
import { App } from './app';
import { AvailabilityService } from './core/services/availability.service';

describe('App', () => {
  beforeEach(async () => {
    localStorage.clear();
    await TestBed.configureTestingModule({
      imports: [App],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter(routes)],
    }).compileComponents();
  });

  afterEach(() => {
    TestBed.inject(HttpTestingController).verify();
    localStorage.clear();
  });

  it('se crea', () => {
    const fixture = TestBed.createComponent(App);
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('no enseña el aviso de servidor caído cuando el servidor responde', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();

    expect((fixture.nativeElement as HTMLElement).querySelector('app-sin-servidor')).toBeNull();
  });

  it('enseña el aviso de servidor caído en cuanto se sabe que no hay servidor', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();

    TestBed.inject(AvailabilityService).noDisponible();
    fixture.detectChanges();

    // El aviso se pone encima de lo que hubiera: puede pasar en cualquier pantalla.
    expect((fixture.nativeElement as HTMLElement).querySelector('app-sin-servidor')).not.toBeNull();
  });
});
