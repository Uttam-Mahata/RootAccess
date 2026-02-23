import { Component, OnInit, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { AdminStateService } from '../../../../services/admin-state';
import { environment } from '../../../../../environments/environment';

@Component({
  selector: 'app-admin-audit',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './admin-audit.html',
  styleUrls: ['./admin-audit.scss']
})
export class AdminAuditComponent implements OnInit {
  private destroyRef = inject(DestroyRef);
  private http = inject(HttpClient);
  adminState = inject(AdminStateService);

  auditLogs: any[] = [];
  auditTotal = 0;
  auditPage = 1;
  isLoadingAudit = false;

  private get apiUrl(): string {
    return environment.apiUrl;
  }

  ngOnInit(): void {
    this.loadAuditLogs();
  }

  loadAuditLogs(): void {
    this.isLoadingAudit = true;
    this.http.get<any>(`${this.apiUrl}/admin/audit-logs?page=${this.auditPage}&limit=50`)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (data) => {
          this.auditLogs = data.logs || [];
          this.auditTotal = data.total || 0;
          this.isLoadingAudit = false;
        },
        error: () => {
          this.auditLogs = [];
          this.isLoadingAudit = false;
          this.adminState.showMessage('Failed to load audit logs', 'error');
        }
      });
  }

  nextAuditPage(): void {
    this.auditPage++;
    this.loadAuditLogs();
  }

  prevAuditPage(): void {
    if (this.auditPage > 1) {
      this.auditPage--;
      this.loadAuditLogs();
    }
  }
}
