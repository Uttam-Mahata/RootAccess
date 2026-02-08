import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { ActivityService, UserActivity } from '../../services/activity';

@Component({
  selector: 'app-activity-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './activity-dashboard.html',
  styleUrls: ['./activity-dashboard.scss']
})
export class ActivityDashboardComponent implements OnInit {
  activity: UserActivity | null = null;
  loading = true;
  error = '';
  categoryKeys: string[] = [];

  constructor(private activityService: ActivityService) {}

  ngOnInit() {
    this.loadActivity();
  }

  loadActivity() {
    this.loading = true;
    this.activityService.getMyActivity().subscribe({
      next: (data) => {
        this.activity = data;
        this.categoryKeys = data.category_progress ? Object.keys(data.category_progress) : [];
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Failed to load activity data';
        this.loading = false;
      }
    });
  }

  getCategoryProgress(key: string): number {
    if (!this.activity?.category_progress?.[key]) return 0;
    const stat = this.activity.category_progress[key];
    return stat.total > 0 ? Math.round((stat.solved / stat.total) * 100) : 0;
  }

  getAchievementIcon(type: string): string {
    const icons: { [key: string]: string } = {
      'first_blood': '🩸',
      'category_master': '🏆',
      'streak': '🔥',
      'speed_demon': '⚡',
      'night_owl': '🦉'
    };
    return icons[type] || '🏅';
  }
}
