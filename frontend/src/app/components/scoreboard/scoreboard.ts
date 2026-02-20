import { Component, OnInit, OnDestroy, ChangeDetectorRef, ChangeDetectionStrategy, ElementRef, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { Subscription } from 'rxjs';
import Chart from 'chart.js/auto';
import { TeamService, Team } from '../../services/team';
import { ScoreboardService } from '../../services/scoreboard';
import { ContestService, ScoreboardContest } from '../../services/contest';
import { WebSocketService } from '../../services/websocket';
import { ThemeService } from '../../services/theme';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../environments/environment';

interface UserScore {
  username: string;
  score: number;
  team_name?: string;
}

interface TeamProgression {
  team_id: string;
  name: string;
  data: Array<{ date: string; score: number }>;
}

@Component({
  selector: 'app-scoreboard',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './scoreboard.html',
  styleUrls: ['./scoreboard.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ScoreboardComponent implements OnInit, OnDestroy {
  teams: Team[] = [];
  users: UserScore[] = [];
  isLoading = true;
  viewMode: 'teams' | 'individuals' = 'teams';
  private wsSub: Subscription | null = null;
  private wsConnectedSub: Subscription | null = null;
  private pollingInterval: any = null;
  private wsConnected = false;

  // Contest state
  contests: ScoreboardContest[] = [];
  selectedContestId: string | null = null;
  contestsLoading = true;

  // Chart data
  teamProgressions: TeamProgression[] = [];
  chartsLoading = false;
  private chartInstances: Chart[] = [];
  showTopTeams = 10;
  private chartsInitialized = false;

  private _teamChartCanvas: ElementRef<HTMLCanvasElement> | undefined;
  @ViewChild('teamChartCanvas') set teamChartCanvas(ref: ElementRef<HTMLCanvasElement> | undefined) {
    this._teamChartCanvas = ref;
    if (ref && this.teamProgressions.length > 0 && !this.chartsInitialized) {
      this.initCharts();
    }
  }

  constructor(
    private teamService: TeamService,
    private scoreboardService: ScoreboardService,
    private contestService: ContestService,
    private wsService: WebSocketService,
    private themeService: ThemeService,
    private http: HttpClient,
    private cdr: ChangeDetectorRef
  ) { }

  ngOnInit(): void {
    this.loadContests();

    // Subscribe to WebSocket scoreboard updates
    this.wsService.connect();

    this.wsConnectedSub = this.wsService.connected$.subscribe(connected => {
      this.wsConnected = connected;
    });

    this.wsSub = this.wsService.on('scoreboard_update').subscribe(() => {
      if (!this.selectedContestId) return;
      if (this.viewMode === 'teams') {
        this.loadTeamScoreboard();
        this.loadTeamStatistics();
      } else {
        this.loadIndividualScoreboard();
      }
    });

    // Polling fallback
    this.pollingInterval = setInterval(() => {
      if (!this.wsConnected && this.selectedContestId) {
        if (this.viewMode === 'teams') {
          this.loadTeamScoreboard();
        } else {
          this.loadIndividualScoreboard();
        }
      }
    }, 10000);
  }

  ngOnDestroy(): void {
    this.wsSub?.unsubscribe();
    this.wsConnectedSub?.unsubscribe();
    if (this.pollingInterval) {
      clearInterval(this.pollingInterval);
    }
    this.destroyCharts();
  }

  trackTeam(_: number, team: any): string {
    return team.team_id ?? team.id;
  }

  trackUser(_: number, user: any): string {
    return user.user_id ?? user.username;
  }

  trackContest(_: number, contest: ScoreboardContest): string {
    return contest.id;
  }

  loadContests(): void {
    this.contestsLoading = true;
    this.contestService.getScoreboardContests().subscribe({
      next: (response) => {
        this.contests = response.contests || [];
        this.contestsLoading = false;

        if (this.contests.length > 0) {
          // Prefer the first running contest; otherwise the first in the list
          const running = this.contests.find(c => c.status === 'running');
          this.selectedContestId = (running || this.contests[0]).id;
          this.loadScoreboardData();
        }
        this.cdr.markForCheck();
      },
      error: () => {
        this.contests = [];
        this.contestsLoading = false;
        this.isLoading = false;
        this.cdr.markForCheck();
      }
    });
  }

  selectContest(contestId: string): void {
    if (this.selectedContestId === contestId) return;
    this.selectedContestId = contestId;
    this.destroyCharts();
    this.loadScoreboardData();
  }

  private loadScoreboardData(): void {
    if (this.viewMode === 'teams') {
      this.loadTeamScoreboard();
      this.loadTeamStatistics();
    } else {
      this.loadIndividualScoreboard();
    }
  }

  getSelectedContest(): ScoreboardContest | undefined {
    return this.contests.find(c => c.id === this.selectedContestId);
  }

  loadTeamStatistics(): void {
    if (!this.selectedContestId) return;
    this.chartsLoading = true;
    this.cdr.markForCheck();

    this.http.get<{ progressions: TeamProgression[] }>(
      `${environment.apiUrl}/scoreboard/teams/statistics?days=30&contest_id=${this.selectedContestId}`,
      { withCredentials: true }
    ).subscribe({
      next: (response) => {
        this.teamProgressions = response.progressions || [];
        this.chartsLoading = false;
        this.cdr.markForCheck();
        if (this._teamChartCanvas && this.teamProgressions.length > 0) {
          this.initCharts();
        }
      },
      error: () => {
        this.chartsLoading = false;
        this.teamProgressions = [];
        this.cdr.markForCheck();
      }
    });
  }

  private destroyCharts(): void {
    this.chartInstances.forEach(chart => chart.destroy());
    this.chartInstances = [];
    this.chartsInitialized = false;
  }

  initCharts(): void {
    if (this.viewMode !== 'teams' || !this.teamProgressions.length) return;

    const solvesEl = this._teamChartCanvas?.nativeElement;
    if (!solvesEl) return;

    this.destroyCharts();
    const isDark = this.themeService.isDarkMode();
    const textColor = isDark ? '#f1f5f9' : '#0f172a';
    const gridColor = isDark ? 'rgba(148,163,184,0.2)' : 'rgba(148,163,184,0.3)';

    const topTeams = this.teamProgressions.slice(0, this.showTopTeams);
    if (topTeams.length === 0) return;

    const colors = [
      '#3b82f6', '#22c55e', '#f59e0b', '#8b5cf6', '#14b8a6',
      '#6366f1', '#ec4899', '#f97316', '#06b6d4', '#84cc16',
      '#0ea5e9', '#a855f7', '#eab308', '#f43f5e', '#10b981'
    ];

    const labels = topTeams[0]?.data?.map(d => d.date.slice(5)) || [];
    if (labels.length === 0) return;

    const datasets = topTeams.map((team, index) => ({
      label: team.name,
      data: team.data?.map(d => d.score) || [],
      borderColor: colors[index % colors.length],
      backgroundColor: colors[index % colors.length] + '20',
      fill: false,
      tension: 0.3,
      borderWidth: 2,
      pointRadius: 2,
      pointHoverRadius: 4
    }));

    try {
      const chart = new Chart(solvesEl, {
        type: 'line',
        data: { labels, datasets },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          interaction: { mode: 'index', intersect: false },
          plugins: {
            legend: {
              position: 'right',
              labels: {
                color: textColor,
                usePointStyle: true,
                padding: 15,
                font: { size: 13, weight: 500, family: 'system-ui, -apple-system, sans-serif' }
              }
            },
            tooltip: {
              backgroundColor: isDark ? 'rgba(30, 41, 59, 0.98)' : 'rgba(255, 255, 255, 0.98)',
              titleColor: textColor,
              bodyColor: textColor,
              borderColor: gridColor,
              borderWidth: 1,
              titleFont: { size: 14, weight: 600 },
              bodyFont: { size: 13, weight: 500 },
              padding: 12
            }
          },
          scales: {
            x: {
              ticks: { color: textColor, maxTicksLimit: 15, font: { size: 12, weight: 500 } },
              grid: { color: gridColor },
              title: { display: false }
            },
            y: {
              beginAtZero: true,
              ticks: { color: textColor, font: { size: 12, weight: 500 } },
              grid: { color: gridColor },
              title: { display: false }
            }
          }
        }
      });

      this.chartInstances.push(chart);
      this.chartsInitialized = true;
    } catch {
      this.chartsInitialized = false;
    }
  }

  switchView(mode: 'teams' | 'individuals'): void {
    if (this.viewMode === mode) return;
    this.viewMode = mode;
    if (!this.selectedContestId) return;

    if (mode === 'teams') {
      this.loadTeamScoreboard();
      if (!this.teamProgressions.length) {
        this.loadTeamStatistics();
      } else {
        this.cdr.markForCheck();
        requestAnimationFrame(() => {
          setTimeout(() => this.initCharts(), 300);
        });
      }
    } else {
      this.loadIndividualScoreboard();
      this.destroyCharts();
    }
  }

  loadTeamScoreboard(): void {
    if (!this.selectedContestId) return;
    this.isLoading = true;
    this.cdr.markForCheck();

    this.teamService.getTeamScoreboard(this.selectedContestId).subscribe({
      next: (response) => {
        this.teams = response.teams || [];
        this.isLoading = false;
        if (this.teamProgressions.length > 0 && this.viewMode === 'teams' && this._teamChartCanvas) {
          this.initCharts();
        }
        this.cdr.markForCheck();
      },
      error: () => {
        this.teams = [];
        this.isLoading = false;
        this.cdr.markForCheck();
      }
    });
  }

  loadIndividualScoreboard(): void {
    if (!this.selectedContestId) return;
    this.isLoading = true;
    this.cdr.markForCheck();

    this.scoreboardService.getScoreboard(this.selectedContestId).subscribe({
      next: (response) => {
        this.users = response || [];
        this.isLoading = false;
        this.cdr.markForCheck();
      },
      error: () => {
        this.users = [];
        this.isLoading = false;
        this.cdr.markForCheck();
      }
    });
  }
}
