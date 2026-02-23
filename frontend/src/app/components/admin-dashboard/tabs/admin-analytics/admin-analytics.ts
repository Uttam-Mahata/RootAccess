import { Component, OnInit, OnDestroy, effect, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import Chart from 'chart.js/auto';
import { AnalyticsService, AdminAnalytics } from '../../../../services/analytics';
import { ThemeService } from '../../../../services/theme';
import { AdminStateService } from '../../../../services/admin-state';

@Component({
  selector: 'app-admin-analytics',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './admin-analytics.html',
  styleUrls: ['./admin-analytics.scss']
})
export class AdminAnalyticsComponent implements OnInit, OnDestroy {
  private destroyRef = inject(DestroyRef);
  private analyticsService = inject(AnalyticsService);
  private themeService = inject(ThemeService);
  adminState = inject(AdminStateService);

  analytics: AdminAnalytics | null = null;
  analyticsLoading = false;
  categoryKeys: string[] = [];
  difficultyKeys: string[] = [];
  private chartInstances: Chart[] = [];

  constructor() {
    effect(() => {
      this.themeService.isDarkMode();
      if (this.analytics && !this.analyticsLoading) {
        setTimeout(() => this.initCharts(), 100);
      }
    });
  }

  ngOnInit(): void {
    this.loadAnalytics();
  }

  ngOnDestroy(): void {
    this.destroyCharts();
  }

  loadAnalytics(): void {
    this.analyticsLoading = true;
    this.destroyCharts();
    this.analyticsService.getPlatformAnalytics().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => {
        this.analytics = data;
        this.categoryKeys = data.category_breakdown ? Object.keys(data.category_breakdown) : [];
        this.difficultyKeys = data.difficulty_breakdown ? Object.keys(data.difficulty_breakdown) : [];
        this.analyticsLoading = false;
        requestAnimationFrame(() => {
          setTimeout(() => this.initCharts(), 100);
        });
      },
      error: () => {
        this.analyticsLoading = false;
        this.adminState.showMessage('Failed to load analytics', 'error');
      }
    });
  }

  private destroyCharts(): void {
    this.chartInstances.forEach(chart => chart.destroy());
    this.chartInstances = [];
  }

  private initCharts(): void {
    if (!this.analytics) return;

    const solvesEl = document.getElementById('solvesOverTimeChart') as HTMLCanvasElement;
    if (!solvesEl) {
      const retryCount = (this as any)._chartRetryCount || 0;
      if (retryCount < 5) {
        (this as any)._chartRetryCount = retryCount + 1;
        setTimeout(() => this.initCharts(), 150);
      } else {
        (this as any)._chartRetryCount = 0;
      }
      return;
    }
    (this as any)._chartRetryCount = 0;

    this.destroyCharts();
    const isDark = this.themeService.isDarkMode();
    const textColor = isDark ? '#f1f5f9' : '#0f172a';
    const gridColor = isDark ? 'rgba(148,163,184,0.2)' : 'rgba(148,163,184,0.3)';

    const commonScaleOpts = (axis: 'x' | 'y', beginAtZero = false, maxTicks?: number) => ({
      beginAtZero: beginAtZero || undefined,
      ticks: { color: textColor, font: { size: 12, weight: 500 as const }, ...(maxTicks ? { maxTicksLimit: maxTicks } : {}) },
      grid: { color: gridColor }
    });
    const commonTooltip = {
      backgroundColor: isDark ? 'rgba(30, 41, 59, 0.98)' : 'rgba(255, 255, 255, 0.98)',
      titleColor: textColor, bodyColor: textColor, borderColor: gridColor, borderWidth: 1,
      titleFont: { size: 14, weight: 600 as const }, bodyFont: { size: 13, weight: 500 as const }, padding: 12
    };

    if (solvesEl && this.analytics.solves_over_time?.length) {
      this.chartInstances.push(new Chart(solvesEl, {
        type: 'line',
        data: {
          labels: this.analytics.solves_over_time.map(e => e.date.slice(5)),
          datasets: [{ label: 'Solves', data: this.analytics.solves_over_time.map(e => e.count), borderColor: '#22c55e', backgroundColor: 'rgba(34,197,94,0.1)', fill: true, tension: 0.3 }]
        },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false }, tooltip: commonTooltip }, scales: { x: commonScaleOpts('x', false, 10), y: commonScaleOpts('y', true) } }
      }));
    }

    const subsEl = document.getElementById('submissionsOverTimeChart') as HTMLCanvasElement;
    if (subsEl && this.analytics.submissions_over_time?.length) {
      this.chartInstances.push(new Chart(subsEl, {
        type: 'line',
        data: {
          labels: this.analytics.submissions_over_time.map(e => e.date.slice(5)),
          datasets: [{ label: 'Submissions', data: this.analytics.submissions_over_time.map(e => e.count), borderColor: '#3b82f6', backgroundColor: 'rgba(59,130,246,0.1)', fill: true, tension: 0.3 }]
        },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false }, tooltip: commonTooltip }, scales: { x: commonScaleOpts('x', false, 10), y: commonScaleOpts('y', true) } }
      }));
    }

    const userGrowthEl = document.getElementById('userGrowthChart') as HTMLCanvasElement;
    if (userGrowthEl && this.analytics.user_growth?.length) {
      this.chartInstances.push(new Chart(userGrowthEl, {
        type: 'line',
        data: {
          labels: this.analytics.user_growth.map(e => e.date.slice(5)),
          datasets: [{ label: 'New users', data: this.analytics.user_growth.map(e => e.count), borderColor: '#8b5cf6', backgroundColor: 'rgba(139,92,246,0.1)', fill: true, tension: 0.3 }]
        },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false }, tooltip: commonTooltip }, scales: { x: commonScaleOpts('x', false, 10), y: commonScaleOpts('y', true) } }
      }));
    }

    const teamGrowthEl = document.getElementById('teamGrowthChart') as HTMLCanvasElement;
    if (teamGrowthEl && this.analytics.team_growth?.length) {
      this.chartInstances.push(new Chart(teamGrowthEl, {
        type: 'line',
        data: {
          labels: this.analytics.team_growth.map(e => e.date.slice(5)),
          datasets: [{ label: 'New teams', data: this.analytics.team_growth.map(e => e.count), borderColor: '#f59e0b', backgroundColor: 'rgba(245,158,11,0.1)', fill: true, tension: 0.3 }]
        },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false }, tooltip: commonTooltip }, scales: { x: commonScaleOpts('x', false, 10), y: commonScaleOpts('y', true) } }
      }));
    }

    const catEl = document.getElementById('categoryBreakdownChart') as HTMLCanvasElement;
    if (catEl && this.categoryKeys.length) {
      const colors = ['#6366f1', '#22c55e', '#eab308', '#ef4444', '#ec4899', '#14b8a6', '#f97316', '#8b5cf6', '#64748b'];
      this.chartInstances.push(new Chart(catEl, {
        type: 'doughnut',
        data: {
          labels: this.categoryKeys.map(k => k.charAt(0).toUpperCase() + k.slice(1)),
          datasets: [{ data: this.categoryKeys.map(k => this.analytics!.category_breakdown[k]), backgroundColor: this.categoryKeys.map((_, i) => colors[i % colors.length]) }]
        },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'right', labels: { color: textColor, font: { size: 13, weight: 500 }, padding: 15 } }, tooltip: commonTooltip } }
      }));
    }

    const diffEl = document.getElementById('difficultyBreakdownChart') as HTMLCanvasElement;
    if (diffEl && this.difficultyKeys.length) {
      const diffColors: Record<string, string> = { easy: '#22c55e', medium: '#eab308', hard: '#ef4444' };
      this.chartInstances.push(new Chart(diffEl, {
        type: 'doughnut',
        data: {
          labels: this.difficultyKeys.map(k => k.charAt(0).toUpperCase() + k.slice(1)),
          datasets: [{ data: this.difficultyKeys.map(k => this.analytics!.difficulty_breakdown[k]), backgroundColor: this.difficultyKeys.map(k => diffColors[k] || '#64748b') }]
        },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'right', labels: { color: textColor, font: { size: 13, weight: 500 }, padding: 15 } }, tooltip: commonTooltip } }
      }));
    }

    const topTeamsEl = document.getElementById('topTeamsChart') as HTMLCanvasElement;
    if (topTeamsEl && this.analytics.top_teams?.length) {
      this.chartInstances.push(new Chart(topTeamsEl, {
        type: 'bar',
        data: {
          labels: this.analytics.top_teams.map(t => t.name.length > 12 ? t.name.slice(0, 12) + '…' : t.name),
          datasets: [{ label: 'Score', data: this.analytics.top_teams.map(t => t.score), backgroundColor: 'rgba(34,197,94,0.7)' }]
        },
        options: { indexAxis: 'y', responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false }, tooltip: commonTooltip }, scales: { x: commonScaleOpts('x', true), y: commonScaleOpts('y') } }
      }));
    }

    const topUsersEl = document.getElementById('topUsersChart') as HTMLCanvasElement;
    if (topUsersEl && this.analytics.top_users?.length) {
      this.chartInstances.push(new Chart(topUsersEl, {
        type: 'bar',
        data: {
          labels: this.analytics.top_users.map(u => u.username.length > 12 ? u.username.slice(0, 12) + '…' : u.username),
          datasets: [{ label: 'Score', data: this.analytics.top_users.map(u => u.score), backgroundColor: 'rgba(59,130,246,0.7)' }]
        },
        options: { indexAxis: 'y', responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false }, tooltip: commonTooltip }, scales: { x: commonScaleOpts('x', true), y: commonScaleOpts('y') } }
      }));
    }

    const popEl = document.getElementById('challengePopularityChart') as HTMLCanvasElement;
    const popData = this.analytics.challenge_popularity?.slice(0, 10) || [];
    if (popEl && popData.length) {
      this.chartInstances.push(new Chart(popEl, {
        type: 'bar',
        data: {
          labels: popData.map(c => c.title.length > 15 ? c.title.slice(0, 15) + '…' : c.title),
          datasets: [{ label: 'Solves', data: popData.map(c => c.solve_count), backgroundColor: 'rgba(139,92,246,0.7)' }]
        },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false }, tooltip: commonTooltip }, scales: { x: { ticks: { color: textColor, maxRotation: 45, font: { size: 12, weight: 500 } }, grid: { color: gridColor } }, y: commonScaleOpts('y', true) } }
      }));
    }
  }
}
