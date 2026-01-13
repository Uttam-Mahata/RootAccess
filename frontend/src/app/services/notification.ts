import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject, interval } from 'rxjs';
import { tap, switchMap, startWith } from 'rxjs/operators';
import { environment } from '../../environments/environment';

export interface Notification {
  id: string;
  title: string;
  content: string;
  type: 'info' | 'warning' | 'success' | 'error';
  created_by: string;
  created_at: string;
  is_active: boolean;
}

export interface CreateNotificationRequest {
  title: string;
  content: string;
  type: string;
}

export interface UpdateNotificationRequest {
  title: string;
  content: string;
  type: string;
  is_active: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class NotificationService {
  private apiUrl = environment.apiUrl;
  private notificationsSubject = new BehaviorSubject<Notification[]>([]);
  private dismissedNotifications = new Set<string>();
  
  notifications$ = this.notificationsSubject.asObservable();

  constructor(private http: HttpClient) {
    // Load dismissed notifications from localStorage
    const dismissed = localStorage.getItem('dismissed_notifications');
    if (dismissed) {
      try {
        const parsed = JSON.parse(dismissed);
        this.dismissedNotifications = new Set(parsed);
      } catch (e) {
        // Ignore parse errors
      }
    }
  }

  // Start polling for notifications (call this from app component)
  startPolling(intervalMs: number = 30000): Observable<Notification[]> {
    return interval(intervalMs).pipe(
      startWith(0),
      switchMap(() => this.getActiveNotifications()),
      tap(notifications => {
        // Filter out dismissed notifications
        const filtered = notifications.filter(n => !this.dismissedNotifications.has(n.id));
        this.notificationsSubject.next(filtered);
      })
    );
  }

  // Get active notifications (public)
  getActiveNotifications(): Observable<Notification[]> {
    return this.http.get<Notification[]>(`${this.apiUrl}/notifications`);
  }

  // Get all notifications (admin)
  getAllNotifications(): Observable<Notification[]> {
    return this.http.get<Notification[]>(`${this.apiUrl}/admin/notifications`);
  }

  // Create notification (admin)
  createNotification(data: CreateNotificationRequest): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/admin/notifications`, data);
  }

  // Update notification (admin)
  updateNotification(id: string, data: UpdateNotificationRequest): Observable<any> {
    return this.http.put<any>(`${this.apiUrl}/admin/notifications/${id}`, data);
  }

  // Delete notification (admin)
  deleteNotification(id: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/admin/notifications/${id}`);
  }

  // Toggle notification active status (admin)
  toggleNotificationActive(id: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/admin/notifications/${id}/toggle`, {});
  }

  // Dismiss a notification (client-side, persisted to localStorage)
  dismissNotification(id: string): void {
    this.dismissedNotifications.add(id);
    localStorage.setItem('dismissed_notifications', JSON.stringify([...this.dismissedNotifications]));
    
    // Update the subject to reflect dismissed notification
    const current = this.notificationsSubject.value;
    this.notificationsSubject.next(current.filter(n => n.id !== id));
  }

  // Clear all dismissed notifications (useful for testing or reset)
  clearDismissed(): void {
    this.dismissedNotifications.clear();
    localStorage.removeItem('dismissed_notifications');
  }
}
