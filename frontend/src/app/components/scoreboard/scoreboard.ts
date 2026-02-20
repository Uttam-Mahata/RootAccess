import { Component, OnInit, OnDestroy, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { Subscription } from 'rxjs';
import Chart from 'chart.js/auto';
import { TeamService, Team } from '../../services/team';
import { ScoreboardService } from '../../services/scoreboard';
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
  styleUrls: ['./scoreboard.scss']
})
export class ScoreboardComponent implements OnInit, OnDestroy {
  teams: Team[] = [];
  users: UserScore[] = [];
  isLoading = true;
  viewMode: 'teams' | 'individuals' = 'teams';
  private wsSub: Subscription | null = null;
  private pollingInterval: any = null;
  
  // Chart data
  teamProgressions: TeamProgression[] = [];
  chartsLoading = false;
  private chartInstances: Chart[] = [];
  showTopTeams = 10; // Show top 10 teams in chart
  private chartsInitialized = false;

  constructor(
    private teamService: TeamService,
    private scoreboardService: ScoreboardService,
    private wsService: WebSocketService,
    private themeService: ThemeService,
    private http: HttpClient,
    private cdr: ChangeDetectorRef
  ) { }

  ngOnInit(): void {
    this.loadTeamScoreboard();
    this.loadTeamStatistics();

    // Subscribe to WebSocket scoreboard updates
    this.wsService.connect();
    this.wsSub = this.wsService.on('scoreboard_update').subscribe(() => {
      console.log('WebSocket scoreboard update received');
      if (this.viewMode === 'teams') {
        this.loadTeamScoreboard();
        this.loadTeamStatistics();
      } else {
        this.loadIndividualScoreboard();
      }
    });

    // Polling fallback: refresh data every 10 seconds
    this.pollingInterval = setInterval(() => {
      if (this.viewMode === 'teams') {
        this.loadTeamScoreboard();
        this.loadTeamStatistics();
      } else {
        this.loadIndividualScoreboard();
      }
    }, 10000); // 10 seconds
  }

  ngOnDestroy(): void {
    this.wsSub?.unsubscribe();
    if (this.pollingInterval) {
      clearInterval(this.pollingInterval);
    }
    this.destroyCharts();
  }

  loadTeamStatistics(): void {
    this.chartsLoading = true;
    console.log('Loading team statistics from:', `${environment.apiUrl}/scoreboard/teams/statistics?days=30`);
    this.http.get<{ progressions: TeamProgression[] }>(`${environment.apiUrl}/scoreboard/teams/statistics?days=30`, { withCredentials: true }).subscribe({
      next: (response) => {
        console.log('Team statistics response:', response);
        this.teamProgressions = response.progressions || [];
        console.log('Loaded progressions:', this.teamProgressions.length, 'teams');
        if (this.teamProgressions.length > 0) {
          console.log('First team data:', this.teamProgressions[0]);
        }
        this.chartsLoading = false;
        this.cdr.markForCheck();
        // Wait for Angular to render the canvas element
        if (this.teamProgressions.length > 0) {
          // Use multiple attempts to ensure canvas is rendered
          let attempts = 0;
          const tryInit = () => {
            attempts++;
            const canvas = document.getElementById('teamScoreProgressionChart') as HTMLCanvasElement;
            if (canvas) {
              console.log('Canvas found, initializing chart (attempt', attempts, ')');
              this.initCharts();
            } else if (attempts < 20) {
              setTimeout(tryInit, 100);
            } else {
              console.error('Failed to find canvas after 20 attempts');
            }
          };
          setTimeout(tryInit, 200);
        } else {
          console.log('No team progressions to display');
        }
      },
      error: (err) => {
        console.error('Error loading team statistics:', err);
        console.error('Error details:', err.error);
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
    console.log('initCharts called', { viewMode: this.viewMode, progressionsLength: this.teamProgressions.length });
    
    if (this.viewMode !== 'teams' || !this.teamProgressions.length) {
      console.log('Charts init skipped:', { viewMode: this.viewMode, progressionsLength: this.teamProgressions.length });
      return;
    }
    
    const solvesEl = document.getElementById('teamScoreProgressionChart') as HTMLCanvasElement;
    console.log('Canvas element:', solvesEl ? 'found' : 'NOT FOUND');
    
    if (!solvesEl) {
      console.log('Canvas element not found, retrying...');
      const retryCount = (this as any)._chartRetryCount || 0;
      if (retryCount < 15) {
        (this as any)._chartRetryCount = retryCount + 1;
        setTimeout(() => this.initCharts(), 200);
      } else {
        (this as any)._chartRetryCount = 0;
        console.error('Failed to find canvas element after 15 retries');
      }
      return;
    }

    // Reset retry counter
    (this as any)._chartRetryCount = 0;

    this.destroyCharts();
    const isDark = this.themeService.isDarkMode();
    const textColor = isDark ? '#f1f5f9' : '#0f172a';
    const gridColor = isDark ? 'rgba(148,163,184,0.2)' : 'rgba(148,163,184,0.3)';

    // Get top N teams
    const topTeams = this.teamProgressions.slice(0, this.showTopTeams);
    if (topTeams.length === 0) {
      console.log('No teams to display in chart');
      return;
    }

    console.log('Top teams for chart:', topTeams.map(t => ({ name: t.name, dataPoints: t.data?.length || 0 })));

    const colors = [
      '#3b82f6', '#22c55e', '#f59e0b', '#8b5cf6', '#14b8a6',
      '#6366f1', '#ec4899', '#f97316', '#06b6d4', '#84cc16',
      '#0ea5e9', '#a855f7', '#eab308', '#f43f5e', '#10b981'
    ];

    // Prepare labels (dates) - use first team's data dates
    const labels = topTeams[0]?.data?.map(d => d.date.slice(5)) || [];
    if (labels.length === 0) {
      console.log('No date labels available, team data:', topTeams[0]);
      return;
    }

    console.log('Chart labels:', labels.length, 'dates');

    // Prepare datasets
    const datasets = topTeams.map((team, index) => {
      const scores = team.data?.map(d => d.score) || [];
      console.log(`Team ${team.name}: ${scores.length} data points, max score: ${Math.max(...scores, 0)}`);
      return {
        label: team.name,
        data: scores,
        borderColor: colors[index % colors.length],
        backgroundColor: colors[index % colors.length] + '20',
        fill: false,
        tension: 0.3,
        borderWidth: 2,
        pointRadius: 2,
        pointHoverRadius: 4
      };
    });

    console.log('Initializing chart with', topTeams.length, 'teams,', labels.length, 'data points');

    try {
      const chart = new Chart(solvesEl, {
        type: 'line',
        data: {
          labels: labels,
          datasets: datasets
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          interaction: {
            mode: 'index',
            intersect: false
          },
          plugins: {
            legend: {
              position: 'right',
              labels: {
                color: textColor,
                usePointStyle: true,
                padding: 15,
                font: {
                  size: 13,
                  weight: 500,
                  family: 'system-ui, -apple-system, sans-serif'
                }
              }
            },
            tooltip: {
              backgroundColor: isDark ? 'rgba(30, 41, 59, 0.98)' : 'rgba(255, 255, 255, 0.98)',
              titleColor: textColor,
              bodyColor: textColor,
              borderColor: gridColor,
              borderWidth: 1,
              titleFont: {
                size: 14,
                weight: 600
              },
              bodyFont: {
                size: 13,
                weight: 500
              },
              padding: 12
            }
          },
          scales: {
            x: {
              ticks: {
                color: textColor,
                maxTicksLimit: 15,
                font: {
                  size: 12,
                  weight: 500
                }
              },
              grid: {
                color: gridColor
              },
              title: {
                display: false
              }
            },
            y: {
              beginAtZero: true,
              ticks: {
                color: textColor,
                font: {
                  size: 12,
                  weight: 500
                }
              },
              grid: {
                color: gridColor
              },
              title: {
                display: false
              }
            }
          }
        }
      });

      this.chartInstances.push(chart);
      this.chartsInitialized = true;
      console.log('✅ Chart initialized successfully!');
    } catch (error) {
      console.error('❌ Error initializing chart:', error);
      this.chartsInitialized = false;
    }
  }

  switchView(mode: 'teams' | 'individuals'): void {
    this.viewMode = mode;
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
    this.isLoading = true;
    this.teamService.getTeamScoreboard().subscribe({
      next: (response) => {
        this.teams = response.teams || [];
        this.isLoading = false;
        // If statistics already loaded, reinitialize charts
        if (this.teamProgressions.length > 0 && this.viewMode === 'teams') {
          setTimeout(() => this.initCharts(), 100);
        }
      },
      error: (err) => {
        console.error(err);
        this.teams = [];
        this.isLoading = false;
      }
    });
  }

  loadIndividualScoreboard(): void {
    this.isLoading = true;
    this.scoreboardService.getScoreboard().subscribe({
      next: (response) => {
        this.users = response || [];
        this.isLoading = false;
      },
      error: (err) => {
        console.error(err);
        this.users = [];
        this.isLoading = false;
      }
    });
  }
}
