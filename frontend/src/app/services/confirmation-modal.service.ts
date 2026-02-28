import { Injectable } from '@angular/core';
import { Subject, Observable } from 'rxjs';

export interface ConfirmationModalData {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  confirmButtonClass?: string;
}

@Injectable({
  providedIn: 'root'
})
export class ConfirmationModalService {
  private modalSubject = new Subject<ConfirmationModalData | null>();
  private resultSubject = new Subject<boolean>();

  show(data: ConfirmationModalData): Observable<boolean> {
    this.modalSubject.next(data);
    return this.resultSubject.asObservable();
  }

  getModalData(): Observable<ConfirmationModalData | null> {
    return this.modalSubject.asObservable();
  }

  confirm(): void {
    this.resultSubject.next(true);
    this.modalSubject.next(null);
  }

  cancel(): void {
    this.resultSubject.next(false);
    this.modalSubject.next(null);
  }
}
