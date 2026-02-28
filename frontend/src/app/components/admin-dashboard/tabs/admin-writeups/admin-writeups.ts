import { Component, OnInit, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { take } from 'rxjs/operators';
import Showdown from 'showdown';
import { AdminTeamService, AdminTeam } from '../../../../services/admin-team';
import { ConfirmationModalService } from '../../../../services/confirmation-modal.service';
import { AdminStateService } from '../../../../services/admin-state';
import { environment } from '../../../../../environments/environment';

@Component({
  selector: 'app-admin-writeups',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './admin-writeups.html',
  styleUrls: ['./admin-writeups.scss']
})
export class AdminWriteupsComponent implements OnInit {
  private destroyRef = inject(DestroyRef);
  private http = inject(HttpClient);
  private adminTeamService = inject(AdminTeamService);
  private confirmationModalService = inject(ConfirmationModalService);
  adminState = inject(AdminStateService);

  writeups: any[] = [];
  isLoadingWriteups = false;
  selectedTeamFilter: string = '';
  selectedWriteup: any | null = null;
  teams: AdminTeam[] = [];

  private showdownConverter = new Showdown.Converter({
    tables: true,
    strikethrough: true,
    tasklists: true,
    smoothLivePreview: true,
    simpleLineBreaks: false,
    openLinksInNewWindow: true,
    emoji: true,
    ghCodeBlocks: true,
    encodeEmails: true,
    simplifiedAutoLink: true,
    literalMidWordUnderscores: true,
    parseImgDimensions: true
  });

  private get apiUrl(): string {
    return environment.apiUrl;
  }

  ngOnInit(): void {
    this.loadTeams();
    this.loadWriteups();
  }

  loadTeams(): void {
    this.adminTeamService.listTeams().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => { this.teams = data || []; },
      error: () => {}
    });
  }

  loadWriteups(): void {
    this.isLoadingWriteups = true;
    const url = this.selectedTeamFilter
      ? `${this.apiUrl}/admin/writeups?team_id=${this.selectedTeamFilter}`
      : `${this.apiUrl}/admin/writeups`;
    this.http.get<any[]>(url).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => {
        this.writeups = data || [];
        this.isLoadingWriteups = false;
        this.adminState.writeupsCount.set(this.writeups.length);
      },
      error: () => {
        this.writeups = [];
        this.isLoadingWriteups = false;
      }
    });
  }

  onTeamFilterChange(): void {
    this.loadWriteups();
  }

  approveWriteup(id: string): void {
    const writeup = this.writeups.find(w => w.id === id);
    if (writeup) {
      const originalStatus = writeup.status;
      writeup.status = 'approved';
      this.http.put<any>(`${this.apiUrl}/admin/writeups/${id}/status`, { status: 'approved' }).subscribe({
        next: () => this.adminState.showMessage('Writeup approved', 'success'),
        error: () => {
          writeup.status = originalStatus;
          this.adminState.showMessage('Error approving writeup', 'error');
        }
      });
    }
  }

  rejectWriteup(id: string): void {
    const writeup = this.writeups.find(w => w.id === id);
    if (writeup) {
      const originalStatus = writeup.status;
      writeup.status = 'rejected';
      this.http.put<any>(`${this.apiUrl}/admin/writeups/${id}/status`, { status: 'rejected' }).subscribe({
        next: () => this.adminState.showMessage('Writeup rejected', 'success'),
        error: () => {
          writeup.status = originalStatus;
          this.adminState.showMessage('Error rejecting writeup', 'error');
        }
      });
    }
  }

  deleteWriteup(id: string): void {
    this.confirmationModalService.show({
      title: 'Delete Writeup',
      message: 'Are you sure you want to delete this writeup?',
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        const index = this.writeups.findIndex(w => w.id === id);
        if (index !== -1) {
          const writeup = this.writeups[index];
          this.writeups.splice(index, 1);
          if (this.selectedWriteup?.id === id) {
            this.selectedWriteup = null;
          }
          this.http.delete<any>(`${this.apiUrl}/admin/writeups/${id}`).subscribe({
            next: () => this.adminState.showMessage('Writeup deleted', 'success'),
            error: () => {
              this.writeups.splice(index, 0, writeup);
              this.adminState.showMessage('Error deleting writeup', 'error');
            }
          });
        }
      }
    });
  }

  viewWriteup(writeup: any): void {
    this.selectedWriteup = writeup;
  }

  renderWriteupContent(writeup: any): string {
    const format = writeup.content_format || 'markdown';
    if (format === 'html') {
      return writeup.content || '';
    } else {
      return this.showdownConverter.makeHtml(writeup.content || '');
    }
  }
}
