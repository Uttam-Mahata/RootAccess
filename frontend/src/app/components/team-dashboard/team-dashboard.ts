import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { RouterModule, Router } from '@angular/router';
import { TeamService, Team, TeamMember, TeamInvitation } from '../../services/team';
import { AuthService } from '../../services/auth';

@Component({
  selector: 'app-team-dashboard',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterModule],
  templateUrl: './team-dashboard.html',
  styleUrls: ['./team-dashboard.scss']
})
export class TeamDashboardComponent implements OnInit {
  teamService = inject(TeamService);
  authService = inject(AuthService);
  
  team: Team | null = null;
  members: TeamMember[] = [];
  pendingInvitations: TeamInvitation[] = [];
  teamPendingInvitations: TeamInvitation[] = [];
  
  // Forms
  createTeamForm: FormGroup;
  inviteForm: FormGroup;
  joinForm: FormGroup;
  
  // UI state
  isLoading = false;
  error = '';
  success = '';
  activeTab: 'info' | 'members' | 'invitations' | 'settings' = 'info';
  inviteMethod: 'username' | 'email' | 'code' = 'username';
  showCreateForm = false;
  showInviteForm = false;
  showJoinForm = false;

  constructor(private fb: FormBuilder, private router: Router) {
    this.createTeamForm = this.fb.group({
      name: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(50)]],
      description: ['', [Validators.maxLength(500)]]
    });

    this.inviteForm = this.fb.group({
      username: [''],
      email: ['', Validators.email]
    });

    this.joinForm = this.fb.group({
      inviteCode: ['', Validators.required]
    });
  }

  ngOnInit(): void {
    this.loadTeamData();
    this.loadPendingInvitations();
  }

  loadTeamData(): void {
    this.isLoading = true;
    this.teamService.getMyTeam().subscribe({
      next: (response) => {
        this.team = response.team || null;
        this.members = response.members || [];
        if (this.team) {
          this.loadTeamPendingInvitations();
        }
        this.isLoading = false;
      },
      error: () => {
        this.team = null;
        this.members = [];
        this.isLoading = false;
      }
    });
  }

  loadPendingInvitations(): void {
    this.teamService.getPendingInvitations().subscribe({
      next: (response) => {
        this.pendingInvitations = response.invitations || [];
      },
      error: () => {
        this.pendingInvitations = [];
      }
    });
  }

  loadTeamPendingInvitations(): void {
    if (!this.team) return;
    this.teamService.getTeamPendingInvitations(this.team.id).subscribe({
      next: (response) => {
        this.teamPendingInvitations = response.invitations || [];
      },
      error: () => {
        this.teamPendingInvitations = [];
      }
    });
  }

  isLeader(): boolean {
    if (!this.team) return false;
    const user = this.authService.getCurrentUser();
    return user !== null && this.team.leader_id === user.id;
  }

  // Create team
  onCreateTeam(): void {
    if (this.createTeamForm.valid && !this.isLoading) {
      this.isLoading = true;
      this.error = '';
      this.success = '';

      const { name, description } = this.createTeamForm.value;
      this.teamService.createTeam(name, description).subscribe({
        next: (response) => {
          this.success = response.message || 'Team created successfully!';
          this.team = response.team || null;
          this.members = response.members || [];
          this.showCreateForm = false;
          this.isLoading = false;
          
          // Load pending invitations for the new team if it exists
          if (this.team) {
            this.loadTeamPendingInvitations();
          }
        },
        error: (err) => {
          this.isLoading = false;
          this.error = err.error?.error || 'Failed to create team';
        }
      });
    }
  }

  // Invite member
  onInviteMember(): void {
    if (this.isLoading) return;
    
    this.isLoading = true;
    this.error = '';
    this.success = '';

    if (!this.team) return;

    if (this.inviteMethod === 'username') {
      const username = this.inviteForm.get('username')?.value;
      if (!username) {
        this.error = 'Username is required';
        this.isLoading = false;
        return;
      }
      this.teamService.inviteByUsername(this.team.id, username).subscribe({
        next: (response) => {
          this.isLoading = false;
          this.success = response.message || 'Invitation sent!';
          this.inviteForm.reset();
          this.showInviteForm = false;
          this.loadTeamPendingInvitations();
        },
        error: (err) => {
          this.isLoading = false;
          this.error = err.error?.error || 'Failed to send invitation';
        }
      });
    } else if (this.inviteMethod === 'email') {
      const email = this.inviteForm.get('email')?.value;
      if (!email) {
        this.error = 'Email is required';
        this.isLoading = false;
        return;
      }
      this.teamService.inviteByEmail(this.team.id, email).subscribe({
        next: (response) => {
          this.isLoading = false;
          this.success = response.message || 'Invitation sent!';
          this.inviteForm.reset();
          this.showInviteForm = false;
          this.loadTeamPendingInvitations();
        },
        error: (err) => {
          this.isLoading = false;
          this.error = err.error?.error || 'Failed to send invitation';
        }
      });
    }
  }

  // Join team by code
  onJoinByCode(): void {
    if (this.joinForm.valid && !this.isLoading) {
      this.isLoading = true;
      this.error = '';
      this.success = '';

      const code = this.joinForm.get('inviteCode')?.value;
      this.teamService.joinByCode(code).subscribe({
        next: (response) => {
          this.success = response.message || 'Joined team successfully!';
          this.team = response.team || null;
          this.members = response.members || [];
          this.showJoinForm = false;
          this.isLoading = false;
          
          // Load pending invitations for the new team if it exists
          if (this.team) {
            this.loadTeamPendingInvitations();
          }
        },
        error: (err) => {
          this.isLoading = false;
          this.error = err.error?.error || 'Failed to join team';
        }
      });
    }
  }

  // Accept invitation
  onAcceptInvitation(invitationId: string): void {
    this.isLoading = true;
    this.error = '';
    this.success = '';
    this.teamService.acceptInvitation(invitationId).subscribe({
      next: (response) => {
        this.success = response.message || 'Joined team!';
        this.team = response.team || null;
        this.members = response.members || [];
        this.isLoading = false;
        
        // Load pending invitations and team invitations
        this.loadPendingInvitations();
        if (this.team) {
          this.loadTeamPendingInvitations();
        }
      },
      error: (err) => {
        this.isLoading = false;
        this.error = err.error?.error || 'Failed to accept invitation';
      }
    });
  }

  // Reject invitation
  onRejectInvitation(invitationId: string): void {
    this.isLoading = true;
    this.error = '';
    this.teamService.rejectInvitation(invitationId).subscribe({
      next: () => {
        this.isLoading = false;
        this.success = 'Invitation rejected';
        this.loadPendingInvitations();
      },
      error: (err) => {
        this.isLoading = false;
        this.error = err.error?.error || 'Failed to reject invitation';
      }
    });
  }

  // Cancel invitation (leader)
  onCancelInvitation(invitationId: string): void {
    if (!this.team) return;
    this.isLoading = true;
    this.error = '';
    this.teamService.cancelInvitation(this.team.id, invitationId).subscribe({
      next: () => {
        this.isLoading = false;
        this.success = 'Invitation cancelled';
        this.loadTeamPendingInvitations();
      },
      error: (err) => {
        this.isLoading = false;
        this.error = err.error?.error || 'Failed to cancel invitation';
      }
    });
  }

  // Remove member (leader)
  onRemoveMember(memberId: string): void {
    if (!this.team || !confirm('Are you sure you want to remove this member?')) return;
    this.isLoading = true;
    this.error = '';
    this.teamService.removeMember(this.team.id, memberId).subscribe({
      next: () => {
        this.isLoading = false;
        this.success = 'Member removed';
        this.loadTeamData();
      },
      error: (err) => {
        this.isLoading = false;
        this.error = err.error?.error || 'Failed to remove member';
      }
    });
  }

  // Leave team
  onLeaveTeam(): void {
    if (!this.team || !confirm('Are you sure you want to leave this team?')) return;
    this.isLoading = true;
    this.error = '';
    this.teamService.leaveTeam(this.team.id).subscribe({
      next: () => {
        this.isLoading = false;
        this.success = 'Left team successfully';
        this.team = null;
        this.members = [];
      },
      error: (err) => {
        this.isLoading = false;
        this.error = err.error?.error || 'Failed to leave team';
      }
    });
  }

  // Delete team
  onDeleteTeam(): void {
    if (!this.team || !confirm('Are you sure you want to delete this team? This action cannot be undone.')) return;
    this.isLoading = true;
    this.error = '';
    this.teamService.deleteTeam(this.team.id).subscribe({
      next: () => {
        this.isLoading = false;
        this.success = 'Team deleted successfully';
        this.team = null;
        this.members = [];
      },
      error: (err) => {
        this.isLoading = false;
        this.error = err.error?.error || 'Failed to delete team';
      }
    });
  }

  // Regenerate invite code
  onRegenerateCode(): void {
    if (!this.team) return;
    this.isLoading = true;
    this.error = '';
    this.teamService.regenerateInviteCode(this.team.id).subscribe({
      next: (response) => {
        this.isLoading = false;
        this.success = 'Invite code regenerated!';
        this.loadTeamData();
      },
      error: (err) => {
        this.isLoading = false;
        this.error = err.error?.error || 'Failed to regenerate code';
      }
    });
  }

  copyInviteCode(): void {
    if (!this.team) return;
    navigator.clipboard.writeText(this.team.invite_code).then(() => {
      this.success = 'Invite code copied to clipboard!';
      setTimeout(() => this.success = '', 3000);
    });
  }

  clearMessages(): void {
    this.error = '';
    this.success = '';
  }
}
