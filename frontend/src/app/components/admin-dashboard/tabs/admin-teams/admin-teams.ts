import { Component, OnInit, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { take } from 'rxjs/operators';
import { AdminTeamService, AdminTeam } from '../../../../services/admin-team';
import { ConfirmationModalService } from '../../../../services/confirmation-modal.service';
import { AdminStateService } from '../../../../services/admin-state';

@Component({
  selector: 'app-admin-teams',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './admin-teams.html',
  styleUrls: ['./admin-teams.scss']
})
export class AdminTeamsComponent implements OnInit {
  private destroyRef = inject(DestroyRef);
  private adminTeamService = inject(AdminTeamService);
  private confirmationModalService = inject(ConfirmationModalService);
  adminState = inject(AdminStateService);

  teams: AdminTeam[] = [];
  teamsLoading = false;
  selectedTeam: AdminTeam | null = null;
  teamScoreDelta: number = 0;
  teamScoreReason: string = '';

  ngOnInit(): void {
    this.loadTeams();
  }

  loadTeams(): void {
    this.teamsLoading = true;
    this.adminTeamService.listTeams().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => {
        this.teams = data || [];
        this.teamsLoading = false;
        this.adminState.teamsCount.set(this.teams.length);
      },
      error: () => {
        this.teamsLoading = false;
        this.adminState.showMessage('Failed to load teams', 'error');
      }
    });
  }

  viewTeamDetails(team: AdminTeam): void {
    this.selectedTeam = this.selectedTeam?.id === team.id ? null : team;
    if (this.selectedTeam) {
      this.teamScoreDelta = 0;
      this.teamScoreReason = '';
    }
  }

  updateTeam(teamId: string, name: string, description: string): void {
    this.adminTeamService.updateTeam(teamId, name, description).subscribe({
      next: () => {
        this.adminState.showMessage('Team updated successfully', 'success');
        this.loadTeams();
      },
      error: () => this.adminState.showMessage('Failed to update team', 'error')
    });
  }

  applyTeamScoreAdjustment(): void {
    if (!this.selectedTeam) return;
    const delta = Number(this.teamScoreDelta);
    if (!delta || isNaN(delta) || delta === 0) {
      this.adminState.showMessage('Please enter a non-zero score delta', 'error');
      return;
    }

    this.adminTeamService.adjustScore(this.selectedTeam.id, delta, this.teamScoreReason || '').subscribe({
      next: () => {
        this.adminState.showMessage('Team score adjusted', 'success');
        this.selectedTeam!.score += delta;
        const idx = this.teams.findIndex(t => t.id === this.selectedTeam!.id);
        if (idx !== -1) {
          this.teams[idx] = { ...this.teams[idx], score: this.teams[idx].score + delta };
        }
        this.teamScoreDelta = 0;
        this.teamScoreReason = '';
      },
      error: () => this.adminState.showMessage('Failed to adjust team score', 'error')
    });
  }

  changeTeamLeader(teamId: string, newLeaderId: string): void {
    this.adminTeamService.updateTeamLeader(teamId, newLeaderId).subscribe({
      next: () => {
        this.adminState.showMessage('Team leader updated successfully', 'success');
        this.loadTeams();
        this.selectedTeam = null;
      },
      error: () => this.adminState.showMessage('Failed to update team leader', 'error')
    });
  }

  removeTeamMember(teamId: string, memberId: string): void {
    this.confirmationModalService.show({
      title: 'Remove Team Member',
      message: 'Are you sure you want to remove this member from the team?',
      confirmText: 'Remove',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        this.adminTeamService.removeMember(teamId, memberId).subscribe({
          next: () => {
            this.adminState.showMessage('Member removed from team', 'success');
            this.loadTeams();
            this.selectedTeam = null;
          },
          error: () => this.adminState.showMessage('Failed to remove member', 'error')
        });
      }
    });
  }

  deleteTeam(teamId: string, teamName: string): void {
    this.confirmationModalService.show({
      title: 'Delete Team',
      message: `Are you sure you want to delete the team "${teamName}"? This action cannot be undone.`,
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        this.adminTeamService.deleteTeam(teamId).subscribe({
          next: () => {
            this.adminState.showMessage('Team deleted successfully', 'success');
            this.loadTeams();
            this.selectedTeam = null;
          },
          error: () => this.adminState.showMessage('Failed to delete team', 'error')
        });
      }
    });
  }
}
