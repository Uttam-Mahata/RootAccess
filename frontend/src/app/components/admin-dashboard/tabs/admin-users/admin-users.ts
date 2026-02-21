import { Component, OnInit, Output, EventEmitter, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { take } from 'rxjs/operators';
import { AdminUserService, AdminUser } from '../../../../services/admin-user';
import { ConfirmationModalService } from '../../../../services/confirmation-modal.service';

@Component({
  selector: 'app-admin-users',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './admin-users.html',
  styleUrls: ['./admin-users.scss']
})
export class AdminUsersComponent implements OnInit {
  private destroyRef = inject(DestroyRef);
  private adminUserService = inject(AdminUserService);
  private confirmationModalService = inject(ConfirmationModalService);

  @Output() countChanged = new EventEmitter<number>();
  @Output() messageEmitted = new EventEmitter<{ msg: string; type: 'success' | 'error' }>();

  users: AdminUser[] = [];
  usersLoading = false;
  selectedUser: AdminUser | null = null;
  userScoreDelta: number = 0;
  userScoreReason: string = '';

  ngOnInit(): void {
    this.loadUsers();
  }

  loadUsers(): void {
    this.usersLoading = true;
    this.adminUserService.listUsers().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => {
        this.users = data || [];
        this.usersLoading = false;
        this.countChanged.emit(this.users.length);
      },
      error: () => {
        this.usersLoading = false;
        this.messageEmitted.emit({ msg: 'Failed to load users', type: 'error' });
      }
    });
  }

  viewUserDetails(user: AdminUser): void {
    this.selectedUser = this.selectedUser?.id === user.id ? null : user;
    if (this.selectedUser) {
      this.userScoreDelta = 0;
      this.userScoreReason = '';
    }
  }

  updateUserStatus(userId: string, status: string): void {
    let reason = '';
    if (status === 'banned') {
      reason = prompt('Enter ban reason:') || '';
    }
    this.adminUserService.updateUserStatus(userId, status, reason).subscribe({
      next: () => {
        this.messageEmitted.emit({ msg: 'User status updated', type: 'success' });
        this.loadUsers();
      },
      error: () => this.messageEmitted.emit({ msg: 'Failed to update user status', type: 'error' })
    });
  }

  updateUserRole(userId: string, role: string): void {
    this.adminUserService.updateUserRole(userId, role).subscribe({
      next: () => {
        this.messageEmitted.emit({ msg: 'User role updated', type: 'success' });
        this.loadUsers();
      },
      error: () => this.messageEmitted.emit({ msg: 'Failed to update user role', type: 'error' })
    });
  }

  deleteUser(userId: string, username: string): void {
    this.confirmationModalService.show({
      title: 'Delete User',
      message: `Are you sure you want to delete the user "${username}"? This action cannot be undone.`,
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        this.adminUserService.deleteUser(userId).subscribe({
          next: () => {
            this.messageEmitted.emit({ msg: 'User deleted successfully', type: 'success' });
            this.loadUsers();
            this.selectedUser = null;
          },
          error: () => this.messageEmitted.emit({ msg: 'Failed to delete user', type: 'error' })
        });
      }
    });
  }

  applyUserScoreAdjustment(): void {
    if (!this.selectedUser) return;
    const delta = Number(this.userScoreDelta);
    if (!delta || isNaN(delta) || delta === 0) {
      this.messageEmitted.emit({ msg: 'Please enter a non-zero score delta', type: 'error' });
      return;
    }

    this.adminUserService.adjustScore(this.selectedUser.id, delta, this.userScoreReason || '').subscribe({
      next: () => {
        this.messageEmitted.emit({ msg: 'User score adjusted', type: 'success' });
        this.userScoreDelta = 0;
        this.userScoreReason = '';
      },
      error: () => this.messageEmitted.emit({ msg: 'Failed to adjust user score', type: 'error' })
    });
  }
}
