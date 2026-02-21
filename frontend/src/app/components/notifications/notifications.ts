import { Component, OnInit, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { NotificationService, Notification } from '../../services/notification';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

@Component({
  selector: 'app-notifications',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './notifications.html',
  styleUrls: ['./notifications.scss']
})
export class NotificationsComponent implements OnInit {
  private destroyRef = inject(DestroyRef);
  notifications: Notification[] = [];
  loading = true;

  constructor(private notificationService: NotificationService) {}

  ngOnInit(): void {
    // Ensure notifications are initialized (HTTP + WebSocket). No-op if already done.
    this.notificationService.initialize();

    // Mark all notifications as read when the page opens (clears badge)
    this.notificationService.markAllAsRead();

    this.notificationService.notifications$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(list => {
      // Sort newest first for a stable, readable view
      this.notifications = [...list].sort(
        (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
      this.loading = false;
    });
  }

  dismiss(id: string): void {
    this.notificationService.dismissNotification(id);
  }

  formatDate(dateStr: string): string {
    if (!dateStr) {
      return '';
    }
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) {
      return dateStr;
    }
    return date.toLocaleString();
  }

  getTypeClasses(type: string): string {
    const base =
      'rounded-xl border shadow-sm px-4 py-3 flex gap-3 items-start bg-white dark:bg-slate-900/80';
    switch (type) {
      case 'info':
        return `${base} border-blue-200 dark:border-blue-500/40`;
      case 'warning':
        return `${base} border-amber-200 dark:border-amber-500/40`;
      case 'success':
        return `${base} border-emerald-200 dark:border-emerald-500/40`;
      case 'error':
        return `${base} border-red-200 dark:border-red-500/40`;
      default:
        return `${base} border-slate-200 dark:border-slate-600/60`;
    }
  }

  getIconPath(type: string): string {
    switch (type) {
      case 'info':
        return 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z';
      case 'warning':
        return 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z';
      case 'success':
        return 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z';
      case 'error':
        return 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z';
      default:
        return 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z';
    }
  }

  getIconColorClasses(type: string): string {
    switch (type) {
      case 'info':
        return 'text-blue-500 dark:text-blue-400';
      case 'warning':
        return 'text-amber-500 dark:text-amber-400';
      case 'success':
        return 'text-emerald-500 dark:text-emerald-400';
      case 'error':
        return 'text-red-500 dark:text-red-400';
      default:
        return 'text-slate-500 dark:text-slate-400';
    }
  }
}

