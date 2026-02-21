import { Component, OnInit, Output, EventEmitter, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { take } from 'rxjs/operators';
import { NotificationService, Notification } from '../../../../services/notification';
import { ConfirmationModalService } from '../../../../services/confirmation-modal.service';

@Component({
  selector: 'app-admin-notifications',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './admin-notifications.html',
  styleUrls: ['./admin-notifications.scss']
})
export class AdminNotificationsComponent implements OnInit {
  private destroyRef = inject(DestroyRef);
  private fb = inject(FormBuilder);
  private notificationService = inject(NotificationService);
  private confirmationModalService = inject(ConfirmationModalService);

  @Output() countChanged = new EventEmitter<number>();
  @Output() messageEmitted = new EventEmitter<{ msg: string; type: 'success' | 'error' }>();

  notificationForm: FormGroup;
  notifications: Notification[] = [];
  isLoadingNotifications = false;
  isEditingNotification = false;
  editingNotificationId: string | null = null;

  notificationTypes = [
    { value: 'info', label: 'Info', colorClass: 'text-blue-500' },
    { value: 'warning', label: 'Warning', colorClass: 'text-amber-500' },
    { value: 'success', label: 'Success', colorClass: 'text-emerald-500' },
    { value: 'error', label: 'Error', colorClass: 'text-red-500' }
  ];

  constructor() {
    this.notificationForm = this.fb.group({
      title: ['', Validators.required],
      content: ['', Validators.required],
      type: ['info', Validators.required]
    });
  }

  ngOnInit(): void {
    this.loadNotifications();
  }

  loadNotifications(): void {
    this.isLoadingNotifications = true;
    this.notificationService.getAllNotifications().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => {
        this.notifications = data || [];
        this.isLoadingNotifications = false;
        this.countChanged.emit(this.notifications.length);
      },
      error: (err) => {
        console.error('Error loading notifications:', err);
        this.notifications = [];
        this.isLoadingNotifications = false;
        this.messageEmitted.emit({ msg: 'Error loading notifications', type: 'error' });
      }
    });
  }

  onSubmitNotification(): void {
    if (this.notificationForm.valid) {
      const formValue = this.notificationForm.value;

      if (this.isEditingNotification && this.editingNotificationId) {
        this.notificationService.updateNotification(this.editingNotificationId, {
          title: formValue.title,
          content: formValue.content,
          type: formValue.type,
          is_active: true
        }).subscribe({
          next: () => {
            this.messageEmitted.emit({ msg: 'Notification updated successfully', type: 'success' });
            this.loadNotifications();
            this.resetNotificationForm();
          },
          error: (err) => {
            console.error('Error updating notification:', err);
            this.messageEmitted.emit({ msg: 'Error updating notification', type: 'error' });
          }
        });
      } else {
        this.notificationService.createNotification({
          title: formValue.title,
          content: formValue.content,
          type: formValue.type
        }).subscribe({
          next: () => {
            this.messageEmitted.emit({ msg: 'Notification created successfully', type: 'success' });
            this.loadNotifications();
            this.resetNotificationForm();
          },
          error: (err) => {
            console.error('Error creating notification:', err);
            this.messageEmitted.emit({ msg: 'Error creating notification', type: 'error' });
          }
        });
      }
    }
  }

  editNotification(notification: Notification): void {
    this.isEditingNotification = true;
    this.editingNotificationId = notification.id;
    this.notificationForm.patchValue({
      title: notification.title,
      content: notification.content,
      type: notification.type
    });
  }

  deleteNotification(notification: Notification): void {
    this.confirmationModalService.show({
      title: 'Delete Notification',
      message: `Are you sure you want to delete the notification "${notification.title}"?`,
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        this.notificationService.deleteNotification(notification.id).subscribe({
          next: () => {
            this.messageEmitted.emit({ msg: 'Notification deleted successfully', type: 'success' });
            this.loadNotifications();
          },
          error: (err) => {
            console.error('Error deleting notification:', err);
            this.messageEmitted.emit({ msg: 'Error deleting notification', type: 'error' });
          }
        });
      }
    });
  }

  toggleNotificationActive(notification: Notification): void {
    this.notificationService.toggleNotificationActive(notification.id).subscribe({
      next: () => {
        this.messageEmitted.emit({
          msg: `Notification ${notification.is_active ? 'deactivated' : 'activated'} successfully`,
          type: 'success'
        });
        this.loadNotifications();
      },
      error: (err) => {
        console.error('Error toggling notification:', err);
        this.messageEmitted.emit({ msg: 'Error toggling notification status', type: 'error' });
      }
    });
  }

  resetNotificationForm(): void {
    this.isEditingNotification = false;
    this.editingNotificationId = null;
    this.notificationForm.reset({ title: '', content: '', type: 'info' });
  }

  cancelEditNotification(): void {
    this.resetNotificationForm();
  }

  getNotificationTypeLabel(value: string): string {
    const type = this.notificationTypes.find(t => t.value === value);
    return type ? type.label : value;
  }

  getNotificationTypeColorClass(value: string): string {
    const type = this.notificationTypes.find(t => t.value === value);
    return type ? type.colorClass : 'text-slate-500';
  }
}
