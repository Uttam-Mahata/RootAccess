import { Component, DestroyRef, inject, OnInit } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { timeout, catchError, of } from 'rxjs';
import { AuthService } from '../../services/auth';
import { TeamService } from '../../services/team';
import { ContestService, Contest } from '../../services/contest';

interface ContestMeta {
  startFormatted: string;
  endFormatted: string;
  relativeTime: string;
}

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
  private destroyRef = inject(DestroyRef);

  upcomingContests: Contest[] = [];
  isLoadingContests = false;
  registrationStatuses: Map<string, boolean> = new Map();
  checkedRegistrationStatuses: Set<string> = new Set();
  checkingRegistrationStatuses: Set<string> = new Set();
  registeringContests: Set<string> = new Set();
  registeredTeamsCounts: Map<string, number> = new Map();
  hasTeam: boolean | null = null; // null = loading, false = no team, true = has team
  contestMeta: Map<string, ContestMeta> = new Map();

  ngOnInit(): void {
    this.teamService.currentTeam$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(team => {
      const hadTeam = this.hasTeam;
      if (team !== null) {
        this.hasTeam = true;
      } else if (this.teamService.teamChecked) {
        this.hasTeam = false; // confirmed: HTTP call completed and user has no team
      }
      // else: keep hasTeam as null (initial BehaviorSubject emission, call still in-flight)
      // If team just became available after contests were already loaded, re-check statuses
      if (!hadTeam && this.hasTeam && this.upcomingContests.length > 0) {
        this.upcomingContests.forEach(contest => {
          // Skip if already confirmed registered — no need to re-query
          if (this.registrationStatuses.get(contest.id) === true) return;
          // Skip if a check is already in flight — avoid race condition with loadUpcomingContests
          if (this.checkingRegistrationStatuses.has(contest.id)) return;
          // Skip if already checked and result was false — will re-check only if not yet resolved
          if (this.checkedRegistrationStatuses.has(contest.id)) return;
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
        // Ensure contests is an array
        this.upcomingContests = contests || [];
        
        // Precompute formatted dates and relative times
        this.contestMeta.clear();
        this.upcomingContests.forEach(contest => {
          this.contestMeta.set(contest.id, {
            startFormatted: this.formatDate(contest.start_time),
            endFormatted: this.formatDate(contest.end_time),
            relativeTime: this.getRelativeTime(contest.start_time),
          });
        });
        
        this.isLoadingContests = false;
        
        // Load registration statuses
        this.upcomingContests.forEach(contest => {
          this.checkRegistrationStatus(contest.id);
        });
      },
      error: (err) => {
        console.error('Error loading upcoming contests:', err);
        this.isLoadingContests = false;
      }
    });
  }

  checkRegistrationStatus(contestId: string): void {
    // Seed from singleton cache for immediate display on navigation/re-render
    if (this.contestService.registrationStatusCache.has(contestId)) {
      this.registrationStatuses.set(contestId, this.contestService.registrationStatusCache.get(contestId)!);
    }

    // Mark as in-flight to allow callers to avoid duplicate requests
    this.checkingRegistrationStatuses.add(contestId);

    // Fallback timeout to ensure we always mark as checked (safety net)
    const fallbackTimeout = setTimeout(() => {
      if (!this.checkedRegistrationStatuses.has(contestId)) {
        console.warn(`Registration status check fallback triggered for contest ${contestId} - marking as checked`);
        this.registrationStatuses.set(contestId, false);
        this.checkedRegistrationStatuses.add(contestId);
        this.checkingRegistrationStatuses.delete(contestId);
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
        this.checkingRegistrationStatuses.delete(contestId);
      },
      error: (err) => {
        // Clear fallback timeout
        clearTimeout(fallbackTimeout);
        // Fallback error handler (shouldn't be reached due to catchError, but just in case)
        console.error('Unexpected error in registration status check (fallback):', err);
        this.registrationStatuses.set(contestId, false);
        this.checkedRegistrationStatuses.add(contestId);
        this.checkingRegistrationStatuses.delete(contestId);
      },
      complete: () => {
        // Clear fallback timeout
        clearTimeout(fallbackTimeout);
        // Ensure we mark as checked even if somehow next wasn't called
        if (!this.checkedRegistrationStatuses.has(contestId)) {
          console.warn(`Registration status check completed but not marked for contest ${contestId} - marking now`);
          this.registrationStatuses.set(contestId, false);
          this.checkedRegistrationStatuses.add(contestId);
          this.checkingRegistrationStatuses.delete(contestId);
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

  private formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  private getRelativeTime(dateStr: string): string {
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
