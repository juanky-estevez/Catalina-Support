import { HttpErrorResponse } from '@angular/common/http';

import { memoriaDelError } from './errores';

describe('memoriaDelError', () => {
  it('conserva las cifras estructuradas de un rechazo de memoria', () => {
    const error = new HttpErrorResponse({
      status: 422,
      error: { requiredBytes: 6 * 1024 ** 3, availableBytes: 5 * 1024 ** 3 },
    });

    expect(memoriaDelError(error)).toEqual({
      requiredBytes: 6 * 1024 ** 3,
      availableBytes: 5 * 1024 ** 3,
    });
  });

  it('rechaza cifras incompletas o inválidas', () => {
    expect(memoriaDelError(new HttpErrorResponse({ error: { requiredBytes: -1 } }))).toBeNull();
    expect(memoriaDelError(new Error('sin respuesta HTTP'))).toBeNull();
  });
});
