import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject } from 'rxjs';
import { tap } from 'rxjs/operators';
import { environment } from '../../environments/environment';

export interface Team {
  id: string;
  name: string;
  description: string;
  leader_id: string;
  member_ids: string[];
  invite_code: string;
  score: number;
  created_at: string;
  updated_at: string;
}

export interface TeamMember {
  id: string;
  username: string;
  is_leader: boolean;
}

export interface TeamInvitation {
  id: string;
  team_id: string;
  team_name: string;
  inviter_id: string;
  inviter_name: string;
  invitee_email?: string;
  invitee_user_id?: string;
  status: string;
  expires_at: string;
  created_at: string;
}

export interface TeamResponse {
  message: string;
  team?: Team;
  members?: TeamMember[];
}

export interface InvitationsResponse {
  invitations: TeamInvitation[];
}

export interface TeamScoreboardResponse {
  teams: Team[];
}

@Injectable({
  providedIn: 'root'
})
export class TeamService {
  private apiUrl = `${environment.apiUrl}/teams`;
  private currentTeamSubject = new BehaviorSubject<Team | null>(null);
  
  currentTeam$ = this.currentTeamSubject.asObservable();

  constructor(private http: HttpClient) {
    // Check team status on init
    this.checkTeamStatus();
  }

  // Team CRUD operations
  createTeam(name: string, description: string): Observable<TeamResponse> {
    return this.http.post<TeamResponse>(`${this.apiUrl}`, { name, description }, { withCredentials: true }).pipe(
      tap(response => {
        if (response.team) {
          this.currentTeamSubject.next(response.team);
        }
      })
    );
  }

  getMyTeam(): Observable<TeamResponse> {
    return this.http.get<TeamResponse>(`${this.apiUrl}/my-team`, { withCredentials: true }).pipe(
      tap(response => {
        if (response.team) {
          this.currentTeamSubject.next(response.team);
        }
      })
    );
  }

  getTeamDetails(teamId: string): Observable<TeamResponse> {
    return this.http.get<TeamResponse>(`${this.apiUrl}/${teamId}`, { withCredentials: true });
  }

  updateTeam(teamId: string, name: string, description: string): Observable<TeamResponse> {
    return this.http.put<TeamResponse>(`${this.apiUrl}/${teamId}`, { name, description }, { withCredentials: true }).pipe(
      tap(response => {
        if (response.team) {
          this.currentTeamSubject.next(response.team);
        }
      })
    );
  }

  deleteTeam(teamId: string): Observable<TeamResponse> {
    return this.http.delete<TeamResponse>(`${this.apiUrl}/${teamId}`, { withCredentials: true }).pipe(
      tap(() => {
        this.currentTeamSubject.next(null);
      })
    );
  }

  // Invitation methods
  inviteByUsername(teamId: string, username: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/${teamId}/invite/username`, { username }, { withCredentials: true });
  }

  inviteByEmail(teamId: string, email: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/${teamId}/invite/email`, { email }, { withCredentials: true });
  }

  joinByCode(code: string): Observable<TeamResponse> {
    return this.http.post<TeamResponse>(`${this.apiUrl}/join/${code}`, {}, { withCredentials: true }).pipe(
      tap(response => {
        if (response.team) {
          this.currentTeamSubject.next(response.team);
        }
      })
    );
  }

  getPendingInvitations(): Observable<InvitationsResponse> {
    return this.http.get<InvitationsResponse>(`${this.apiUrl}/invitations`, { withCredentials: true });
  }

  acceptInvitation(invitationId: string): Observable<TeamResponse> {
    return this.http.post<TeamResponse>(`${this.apiUrl}/invitations/${invitationId}/accept`, {}, { withCredentials: true }).pipe(
      tap(response => {
        if (response.team) {
          this.currentTeamSubject.next(response.team);
        }
      })
    );
  }

  rejectInvitation(invitationId: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/invitations/${invitationId}/reject`, {}, { withCredentials: true });
  }

  // Team invitation management (for leaders)
  getTeamPendingInvitations(teamId: string): Observable<InvitationsResponse> {
    return this.http.get<InvitationsResponse>(`${this.apiUrl}/${teamId}/invitations`, { withCredentials: true });
  }

  cancelInvitation(teamId: string, invitationId: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/${teamId}/invitations/${invitationId}`, { withCredentials: true });
  }

  // Member management
  removeMember(teamId: string, userId: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/${teamId}/members/${userId}`, { withCredentials: true });
  }

  leaveTeam(teamId: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/${teamId}/leave`, {}, { withCredentials: true }).pipe(
      tap(() => {
        this.currentTeamSubject.next(null);
      })
    );
  }

  regenerateInviteCode(teamId: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/${teamId}/regenerate-code`, {}, { withCredentials: true });
  }

  // Scoreboard
  getTeamScoreboard(): Observable<TeamScoreboardResponse> {
    return this.http.get<TeamScoreboardResponse>(`${environment.apiUrl}/scoreboard/teams`, { withCredentials: true });
  }

  // Helper methods
  getCurrentTeam(): Team | null {
    return this.currentTeamSubject.value;
  }

  isInTeam(): boolean {
    return this.currentTeamSubject.value !== null;
  }

  isTeamLeader(): boolean {
    // This requires user ID from auth service
    return false; // Will be implemented with proper check
  }

  checkTeamStatus(): void {
    this.http.get<TeamResponse>(`${this.apiUrl}/my-team`, { withCredentials: true }).subscribe({
      next: (response) => {
        if (response.team) {
          this.currentTeamSubject.next(response.team);
        } else {
          this.currentTeamSubject.next(null);
        }
      },
      error: () => {
        this.currentTeamSubject.next(null);
      }
    });
  }
}
