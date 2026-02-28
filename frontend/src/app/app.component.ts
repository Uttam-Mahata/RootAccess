import { Component, inject, OnInit, OnDestroy } from '@angular/core';
import { Router, RouterOutlet, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { AuthService } from './services/auth';
import { NotificationService, Notification } from './services/notification';
import { NotificationBannerComponent } from './components/notification-banner/notification-banner';
import { ConfirmationModalComponent } from './components/confirmation-modal/confirmation-modal';
import { Subscription } from 'rxjs';
import 'zone.js';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterModule, CommonModule, NotificationBannerComponent, ConfirmationModalComponent],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit, OnDestroy {
  public authService = inject(AuthService);
  public notificationService = inject(NotificationService);
  private router = inject(Router);

  notifications: Notification[] = [];
  unreadCount = 0;
  showNotificationDropdown = false;
  showMobileMenu = false;
  private notificationSub?: Subscription;
  private unreadCountSub?: Subscription;

  ngOnInit(): void {
    this.notificationSub = this.notificationService.notifications$.subscribe(
      notifications => this.notifications = notifications
    );
    this.unreadCountSub = this.notificationService.unreadCount$.subscribe(
      count => this.unreadCount = count
    );
  }

  ngOnDestroy(): void {
    this.notificationSub?.unsubscribe();
    this.unreadCountSub?.unsubscribe();
  }

  isLoggedIn(): boolean {
    return this.authService.isLoggedIn();
  }

  isAdmin(): boolean {
    return this.authService.isAdmin();
  }

  logout(): void {
    this.authService.logout().subscribe(() => {
      window.location.href = '/login';
    });
  }

  toggleMobileMenu(): void {
    this.showMobileMenu = !this.showMobileMenu;
    // Close notification dropdown when opening mobile menu
    if (this.showMobileMenu) {
      this.showNotificationDropdown = false;
    }
  }

  closeMobileMenu(): void {
    this.showMobileMenu = false;
  }

  toggleNotificationDropdown(): void {
    this.showNotificationDropdown = !this.showNotificationDropdown;
    // Close mobile menu when opening notifications
    if (this.showNotificationDropdown) {
      this.showMobileMenu = false;
    }
  }

  closeNotificationDropdown(): void {
    this.showNotificationDropdown = false;
  }

  openNotificationsFromDropdown(): void {
    this.closeNotificationDropdown();
    this.router.navigate(['/notifications']);
  }

  dismissNotification(id: string, event: Event): void {
    event.stopPropagation();
    this.notificationService.dismissNotification(id);
  }

  getNotificationIcon(type: string): string {
    switch (type) {
      case 'info': return 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z';
      case 'warning': return 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z';
      case 'success': return 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z';
      case 'error': return 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z';
      default: return 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z';
    }
  }

  getNotificationColor(type: string): string {
    switch (type) {
      case 'info': return 'text-blue-500';
      case 'warning': return 'text-amber-500';
      case 'success': return 'text-emerald-500';
      case 'error': return 'text-red-500';
      default: return 'text-slate-500';
    }
  }
}
