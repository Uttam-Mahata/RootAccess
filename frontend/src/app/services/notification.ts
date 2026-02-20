import { Injectable, OnDestroy } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject, Subscription } from 'rxjs';
import { environment } from '../../environments/environment';
import { WebSocketService } from './websocket';

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
export class NotificationService implements OnDestroy {
  private apiUrl = environment.apiUrl;
  private notificationsSubject = new BehaviorSubject<Notification[]>([]);
  private dismissedNotifications = new Set<string>();
  private wsSubscriptions: Subscription[] = [];
  private initialized = false;

  notifications$ = this.notificationsSubject.asObservable();

  constructor(private http: HttpClient, private wsService: WebSocketService) {
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

  /**
   * Load active notifications once via HTTP, then keep the list up to date
   * through WebSocket push events. Call this once at app startup.
   */
  initialize(): void {
    if (this.initialized) return;
    this.initialized = true;

    // Ensure WebSocket connection is open
    this.wsService.connect();

    // Initial HTTP load
    this.getActiveNotifications().subscribe({
      next: (notifications) => this.setNotifications(notifications),
      error: (err) => console.error('Failed to load notifications:', err)
    });

    // Push: new active notification created
    this.wsSubscriptions.push(
      this.wsService.on('notification:created').subscribe((notification: Notification) => {
        if (!notification.is_active) return;
        const current = this.notificationsSubject.value;
        if (!current.find(n => n.id === notification.id) && !this.dismissedNotifications.has(notification.id)) {
          this.notificationsSubject.next([...current, notification]);
        }
      })
    );

    // Push: notification updated (active state may have changed)
    this.wsSubscriptions.push(
      this.wsService.on('notification:updated').subscribe((notification: Notification) => {
        const current = this.notificationsSubject.value;
        if (!notification.is_active || this.dismissedNotifications.has(notification.id)) {
          // Remove from visible list if deactivated or dismissed
          this.notificationsSubject.next(current.filter(n => n.id !== notification.id));
        } else {
          const idx = current.findIndex(n => n.id === notification.id);
          if (idx >= 0) {
            // Replace in-place
            const updated = [...current];
            updated[idx] = notification;
            this.notificationsSubject.next(updated);
          } else {
            this.notificationsSubject.next([...current, notification]);
          }
        }
      })
    );

    // Push: notification deleted
    this.wsSubscriptions.push(
      this.wsService.on('notification:deleted').subscribe((payload: { id: string }) => {
        const current = this.notificationsSubject.value;
        this.notificationsSubject.next(current.filter(n => n.id !== payload.id));
      })
    );
  }

  private setNotifications(notifications: Notification[]): void {
    const filtered = notifications.filter(n => !this.dismissedNotifications.has(n.id));
    this.notificationsSubject.next(filtered);
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

  ngOnDestroy(): void {
    this.wsSubscriptions.forEach(s => s.unsubscribe());
    this.wsSubscriptions = [];
  }
}
