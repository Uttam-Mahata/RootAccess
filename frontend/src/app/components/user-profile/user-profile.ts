import { Component, OnInit, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { ProfileService, UserProfile, CategoryStats } from '../../services/profile';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

@Component({
  selector: 'app-user-profile',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './user-profile.html',
  styleUrls: ['./user-profile.scss']
})
export class UserProfileComponent implements OnInit {
  private destroyRef = inject(DestroyRef);
  profile: UserProfile | null = null;
  isLoading = true;
  error = '';

  constructor(
    private route: ActivatedRoute,
    private profileService: ProfileService
  ) {}

  ngOnInit(): void {
    const username = this.route.snapshot.paramMap.get('username');
    if (username) {
      this.loadProfile(username);
    } else {
      this.error = 'No username provided';
      this.isLoading = false;
    }
  }

  loadProfile(username: string): void {
    this.isLoading = true;
    this.profileService.getUserProfile(username).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (profile) => {
        this.profile = profile;
        this.isLoading = false;
      },
      error: (err) => {
        console.error('Error loading profile:', err);
        this.error = 'User not found or profile is private';
        this.isLoading = false;
      }
    });
  }

  getCategoryColor(category: string): string {
    const colors: { [key: string]: string } = {
      'web': 'from-blue-500 to-blue-600',
      'crypto': 'from-purple-500 to-purple-600',
      'pwn': 'from-red-500 to-red-600',
      'reverse': 'from-amber-500 to-amber-600',
      'forensics': 'from-emerald-500 to-emerald-600',
      'networking': 'from-cyan-500 to-cyan-600',
      'steganography': 'from-pink-500 to-pink-600',
      'osint': 'from-indigo-500 to-indigo-600',
      'misc': 'from-slate-500 to-slate-600'
    };
    return colors[category.toLowerCase()] || 'from-slate-500 to-slate-600';
  }

  getCategoryLabel(category: string): string {
    const labels: { [key: string]: string } = {
      'web': 'Web Exploitation',
      'crypto': 'Cryptography',
      'pwn': 'Binary Exploitation',
      'reverse': 'Reverse Engineering',
      'forensics': 'Digital Forensics',
      'networking': 'Networking',
      'steganography': 'Steganography',
      'osint': 'OSINT',
      'misc': 'General Skills'
    };
    return labels[category.toLowerCase()] || category;
  }

  getDifficultyColor(difficulty: string): string {
    const colors: { [key: string]: string } = {
      'easy': 'text-emerald-600 dark:text-emerald-400',
      'medium': 'text-amber-600 dark:text-amber-400',
      'hard': 'text-red-600 dark:text-red-400'
    };
    return colors[difficulty.toLowerCase()] || 'text-slate-600 dark:text-slate-400';
  }

  getDifficultyBgColor(difficulty: string): string {
    const colors: { [key: string]: string } = {
      'easy': 'bg-emerald-100 dark:bg-emerald-500/20',
      'medium': 'bg-amber-100 dark:bg-amber-500/20',
      'hard': 'bg-red-100 dark:bg-red-500/20'
    };
    return colors[difficulty.toLowerCase()] || 'bg-slate-100 dark:bg-slate-500/20';
  }

  getMaxCategoryPoints(): number {
    if (!this.profile?.category_stats) return 100;
    return Math.max(...this.profile.category_stats.map(s => s.total_points), 100);
  }

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
      year: 'numeric', 
      month: 'short', 
      day: 'numeric' 
    });
  }
}
