import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NotificationService, Notification } from '../../services/notification';
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-notification-banner',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './notification-banner.html',
  styleUrls: ['./notification-banner.scss']
})
export class NotificationBannerComponent implements OnInit, OnDestroy {
  notifications: Notification[] = [];
  private subscription?: Subscription;

  constructor(private notificationService: NotificationService) {}

  ngOnInit(): void {
    // Initialize once: HTTP fetch + WebSocket push listeners (no polling)
    this.notificationService.initialize();

    this.subscription = this.notificationService.notifications$.subscribe(notifications => {
      this.notifications = notifications;
    });
  }

  ngOnDestroy(): void {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
  }

  dismiss(id: string): void {
    this.notificationService.dismissNotification(id);
  }

  getNotificationClasses(type: string): string {
    const baseClasses = 'relative px-4 py-3 rounded-lg border shadow-lg backdrop-blur-sm';
    
    switch (type) {
      case 'info':
        return `${baseClasses} bg-blue-50/90 dark:bg-blue-500/10 border-blue-200 dark:border-blue-500/30 text-blue-800 dark:text-blue-300`;
      case 'warning':
        return `${baseClasses} bg-amber-50/90 dark:bg-amber-500/10 border-amber-200 dark:border-amber-500/30 text-amber-800 dark:text-amber-300`;
      case 'success':
        return `${baseClasses} bg-emerald-50/90 dark:bg-emerald-500/10 border-emerald-200 dark:border-emerald-500/30 text-emerald-800 dark:text-emerald-300`;
      case 'error':
        return `${baseClasses} bg-red-50/90 dark:bg-red-500/10 border-red-200 dark:border-red-500/30 text-red-800 dark:text-red-300`;
      default:
        return `${baseClasses} bg-slate-50/90 dark:bg-slate-500/10 border-slate-200 dark:border-slate-500/30 text-slate-800 dark:text-slate-300`;
    }
  }

  getIconClasses(type: string): string {
    switch (type) {
      case 'info':
        return 'text-blue-600 dark:text-blue-400';
      case 'warning':
        return 'text-amber-600 dark:text-amber-400';
      case 'success':
        return 'text-emerald-600 dark:text-emerald-400';
      case 'error':
        return 'text-red-600 dark:text-red-400';
      default:
        return 'text-slate-600 dark:text-slate-400';
    }
  }
}
