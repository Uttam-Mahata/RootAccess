import { Component, OnInit, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, ActivatedRoute, Router, NavigationEnd } from '@angular/router';
import { AdminStateService } from '../../services/admin-state';
import { filter } from 'rxjs/operators';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

@Component({
  selector: 'app-admin-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule
  ],
  templateUrl: './admin-dashboard.html',
  styleUrls: ['./admin-dashboard.scss']
})
export class AdminDashboardComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private destroyRef = inject(DestroyRef);
  router = inject(Router);
  adminState = inject(AdminStateService);

  sidebarOpen = true;
  mobileMenuOpen = false;

  constructor() {
    this.router.events.pipe(
      filter(event => event instanceof NavigationEnd),
      takeUntilDestroyed()
    ).subscribe(() => {
      if (!this.router.url.includes('/challenges')) {
        this.adminState.isEditMode.set(false);
      }
    });
  }

  ngOnInit(): void {
    const params = this.route.snapshot.queryParams;
    const tab = params['tab'] as string;

    if (tab === 'create') {
      this.adminState.challengesInitialView.set('create');
    } else if (tab === 'manage') {
      this.adminState.challengesInitialView.set('manage');
    }
  }

  showCreateChallenge(): void {
    this.adminState.challengesInitialView.set('create');
    this.mobileMenuOpen = false;
  }

  toggleSidebar(): void {
    this.sidebarOpen = !this.sidebarOpen;
  }

  toggleMobileMenu(): void {
    this.mobileMenuOpen = !this.mobileMenuOpen;
  }
}
