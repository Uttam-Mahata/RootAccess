import { HttpInterceptorFn } from '@angular/common/http';

export const credentialsInterceptor: HttpInterceptorFn = (req, next) => {
  // Clone the request to add withCredentials
  const clonedRequest = req.clone({
    withCredentials: true
  });

  return next(clonedRequest);
};
