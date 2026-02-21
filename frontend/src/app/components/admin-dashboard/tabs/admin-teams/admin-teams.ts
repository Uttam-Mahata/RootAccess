import { Component, OnInit, Output, EventEmitter, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { take } from 'rxjs/operators';
import { AdminTeamService, AdminTeam } from '../../../../services/admin-team';
import { ConfirmationModalService } from '../../../../services/confirmation-modal.service';

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

  @Output() countChanged = new EventEmitter<number>();
  @Output() messageEmitted = new EventEmitter<{ msg: string; type: 'success' | 'error' }>();

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
        this.countChanged.emit(this.teams.length);
      },
      error: () => {
        this.teamsLoading = false;
        this.messageEmitted.emit({ msg: 'Failed to load teams', type: 'error' });
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
        this.messageEmitted.emit({ msg: 'Team updated successfully', type: 'success' });
        this.loadTeams();
      },
      error: () => this.messageEmitted.emit({ msg: 'Failed to update team', type: 'error' })
    });
  }

  applyTeamScoreAdjustment(): void {
    if (!this.selectedTeam) return;
    const delta = Number(this.teamScoreDelta);
    if (!delta || isNaN(delta) || delta === 0) {
      this.messageEmitted.emit({ msg: 'Please enter a non-zero score delta', type: 'error' });
      return;
    }

    this.adminTeamService.adjustScore(this.selectedTeam.id, delta, this.teamScoreReason || '').subscribe({
      next: () => {
        this.messageEmitted.emit({ msg: 'Team score adjusted', type: 'success' });
        this.selectedTeam!.score += delta;
        const idx = this.teams.findIndex(t => t.id === this.selectedTeam!.id);
        if (idx !== -1) {
          this.teams[idx] = { ...this.teams[idx], score: this.teams[idx].score + delta };
        }
        this.teamScoreDelta = 0;
        this.teamScoreReason = '';
      },
      error: () => this.messageEmitted.emit({ msg: 'Failed to adjust team score', type: 'error' })
    });
  }

  changeTeamLeader(teamId: string, newLeaderId: string): void {
    this.adminTeamService.updateTeamLeader(teamId, newLeaderId).subscribe({
      next: () => {
        this.messageEmitted.emit({ msg: 'Team leader updated successfully', type: 'success' });
        this.loadTeams();
        this.selectedTeam = null;
      },
      error: () => this.messageEmitted.emit({ msg: 'Failed to update team leader', type: 'error' })
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
            this.messageEmitted.emit({ msg: 'Member removed from team', type: 'success' });
            this.loadTeams();
            this.selectedTeam = null;
          },
          error: () => this.messageEmitted.emit({ msg: 'Failed to remove member', type: 'error' })
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
            this.messageEmitted.emit({ msg: 'Team deleted successfully', type: 'success' });
            this.loadTeams();
            this.selectedTeam = null;
          },
          error: () => this.messageEmitted.emit({ msg: 'Failed to delete team', type: 'error' })
        });
      }
    });
  }
}
