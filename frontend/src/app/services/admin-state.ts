import { Injectable, signal } from '@angular/core';

export interface AdminMessage {
  msg: string;
  type: 'success' | 'error';
}

@Injectable({
  providedIn: 'root'
})
export class AdminStateService {
  challengeCount = signal(0);
  notificationCount = signal(0);
  writeupsCount = signal(0);
  usersCount = signal(0);
  teamsCount = signal(0);

  message = signal<AdminMessage | null>(null);

  isEditMode = signal(false);
  challengesInitialView = signal<'create' | 'manage'>('manage');

  showMessage(msg: string, type: 'success' | 'error'): void {
    this.message.set({ msg, type });
    if (type === 'success') {
      setTimeout(() => {
        const current = this.message();
        if (current && current.msg === msg) {
          this.message.set(null);
        }
      }, 5000);
    }
  }

  clearMessage(): void {
    this.message.set(null);
  }
}
