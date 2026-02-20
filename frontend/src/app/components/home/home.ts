import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { timeout, catchError, of } from 'rxjs';
import { AuthService } from '../../services/auth';
import { TeamService } from '../../services/team';
import { ContestService, Contest } from '../../services/contest';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './home.html',
  styleUrls: ['./home.scss']
})
export class HomeComponent implements OnInit {
  authService = inject(AuthService);
  teamService = inject(TeamService);
  contestService = inject(ContestService);

  upcomingContests: Contest[] = [];
  isLoadingContests = false;
  registrationStatuses: Map<string, boolean> = new Map();
  checkedRegistrationStatuses: Set<string> = new Set();
  registeringContests: Set<string> = new Set();
  registeredTeamsCounts: Map<string, number> = new Map();
  hasTeam = false;

  ngOnInit(): void {
    this.teamService.currentTeam$.subscribe(team => {
      const hadTeam = this.hasTeam;
      this.hasTeam = team !== null;
      // If team just became available after contests were already loaded, re-check statuses
      if (!hadTeam && this.hasTeam && this.upcomingContests.length > 0) {
        this.upcomingContests.forEach(contest => {
          this.checkRegistrationStatus(contest.id);
        });
      }
    });
    this.loadUpcomingContests();
  }

  loadUpcomingContests(): void {
    this.isLoadingContests = true;
    this.contestService.getUpcomingContests().subscribe({
      next: (contests) => {
        this.upcomingContests = contests;
        this.isLoadingContests = false;
        // Load registration statuses for each contest
        contests.forEach(contest => {
          this.checkRegistrationStatus(contest.id);
        });
      },
      error: () => {
        this.isLoadingContests = false;
      }
    });
  }

  checkRegistrationStatus(contestId: string): void {
    // Fallback timeout to ensure we always mark as checked (safety net)
    const fallbackTimeout = setTimeout(() => {
      if (!this.checkedRegistrationStatuses.has(contestId)) {
        console.warn(`Registration status check fallback triggered for contest ${contestId} - marking as checked`);
        this.registrationStatuses.set(contestId, false);
        this.checkedRegistrationStatuses.add(contestId);
      }
    }, 12000); // 12 seconds (longer than the 10s timeout)

    // Always check registration status - backend returns {"registered": false} if no team
    // Add timeout to prevent infinite loading if request hangs
    this.contestService.getRegistrationStatus(contestId).pipe(
      timeout(10000), // 10 second timeout
      catchError((err) => {
        // On error (401, 500, network error, timeout, etc.), return default value to prevent infinite loading
        console.error(`Failed to check registration status for contest ${contestId}:`, err);
        // Always return a value so the observable completes and next() is called
        return of({ registered: false });
      })
    ).subscribe({
      next: (status) => {
        // Clear fallback timeout since we got a response
        clearTimeout(fallbackTimeout);
        // Always mark as checked, whether successful or errored
        this.registrationStatuses.set(contestId, status?.registered ?? false);
        this.checkedRegistrationStatuses.add(contestId);
      },
      error: (err) => {
        // Clear fallback timeout
        clearTimeout(fallbackTimeout);
        // Fallback error handler (shouldn't be reached due to catchError, but just in case)
        console.error('Unexpected error in registration status check (fallback):', err);
        this.registrationStatuses.set(contestId, false);
        this.checkedRegistrationStatuses.add(contestId);
      },
      complete: () => {
        // Clear fallback timeout
        clearTimeout(fallbackTimeout);
        // Ensure we mark as checked even if somehow next wasn't called
        if (!this.checkedRegistrationStatuses.has(contestId)) {
          console.warn(`Registration status check completed but not marked for contest ${contestId} - marking now`);
          this.registrationStatuses.set(contestId, false);
          this.checkedRegistrationStatuses.add(contestId);
        }
      }
    });
    this.contestService.getRegisteredTeamsCount(contestId).subscribe({
      next: (resp) => this.registeredTeamsCounts.set(contestId, resp.count),
      error: () => {}
    });
  }

  isRegistered(contestId: string): boolean {
    return this.registrationStatuses.get(contestId) || false;
  }

  hasCheckedRegistrationStatus(contestId: string): boolean {
    return this.checkedRegistrationStatuses.has(contestId);
  }

  registerForContest(contestId: string): void {
    if (this.registeringContests.has(contestId)) return;
    
    this.registeringContests.add(contestId);
    this.contestService.registerTeamForContest(contestId).subscribe({
      next: () => {
        this.registrationStatuses.set(contestId, true);
        this.checkedRegistrationStatuses.add(contestId);
        this.registeringContests.delete(contestId);
      },
      error: (err) => {
        alert(err.error?.error || 'Failed to register for contest');
        this.registeringContests.delete(contestId);
      }
    });
  }

  unregisterFromContest(contestId: string): void {
    if (this.registeringContests.has(contestId)) return;
    
    this.registeringContests.add(contestId);
    this.contestService.unregisterTeamFromContest(contestId).subscribe({
      next: () => {
        this.registrationStatuses.set(contestId, false);
        this.checkedRegistrationStatuses.add(contestId);
        this.registeringContests.delete(contestId);
      },
      error: (err) => {
        alert(err.error?.error || 'Failed to unregister from contest');
        this.registeringContests.delete(contestId);
      }
    });
  }

  formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  getRelativeTime(dateStr: string): string {
    const now = new Date();
    const target = new Date(dateStr);
    const diffMs = target.getTime() - now.getTime();
    if (diffMs <= 0) return 'Started';

    const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    const hours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

    if (days > 0) return `Starts in ${days}d ${hours}h`;
    if (hours > 0) return `Starts in ${hours}h ${minutes}m`;
    return `Starts in ${minutes}m`;
  }
}
