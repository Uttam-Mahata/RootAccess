import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';

import { AdminChallengesComponent } from './tabs/admin-challenges/admin-challenges';
import { AdminNotificationsComponent } from './tabs/admin-notifications/admin-notifications';
import { AdminContestComponent } from './tabs/admin-contest/admin-contest';
import { AdminWriteupsComponent } from './tabs/admin-writeups/admin-writeups';
import { AdminAuditComponent } from './tabs/admin-audit/admin-audit';
import { AdminAnalyticsComponent } from './tabs/admin-analytics/admin-analytics';
import { AdminUsersComponent } from './tabs/admin-users/admin-users';
import { AdminTeamsComponent } from './tabs/admin-teams/admin-teams';

type AdminTab = 'challenges' | 'notifications' | 'contest' | 'writeups' | 'audit' | 'analytics' | 'users' | 'teams';

@Component({
  selector: 'app-admin-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    AdminChallengesComponent,
    AdminNotificationsComponent,
    AdminContestComponent,
    AdminWriteupsComponent,
    AdminAuditComponent,
    AdminAnalyticsComponent,
    AdminUsersComponent,
    AdminTeamsComponent
  ],
  templateUrl: './admin-dashboard.html',
  styleUrls: ['./admin-dashboard.scss']
})
export class AdminDashboardComponent implements OnInit {
  private route = inject(ActivatedRoute);

  activeTab: AdminTab = 'challenges';
  challengesInitialView: 'create' | 'manage' = 'manage';

  sidebarOpen = true;
  mobileMenuOpen = false;

  isEditMode = false;

  message = '';
  messageType: 'success' | 'error' = 'success';

  // Badge counts — set by sub-components via @Output
  challengeCount = 0;
  notificationCount = 0;
  writeupsCount = 0;
  usersCount = 0;
  teamsCount = 0;

  ngOnInit(): void {
    const params = this.route.snapshot.queryParams;
    const tab = params['tab'] as string;

    if (tab === 'create') {
      this.activeTab = 'challenges';
      this.challengesInitialView = 'create';
    } else if (tab === 'manage') {
      this.activeTab = 'challenges';
      this.challengesInitialView = 'manage';
    } else if (tab && ['challenges', 'notifications', 'contest', 'writeups', 'audit', 'analytics', 'users', 'teams'].includes(tab)) {
      this.activeTab = tab as AdminTab;
    } else {
      this.activeTab = 'challenges';
    }
  }

  switchTab(tab: AdminTab): void {
    if (this.activeTab === tab) return;
    this.activeTab = tab;
    this.mobileMenuOpen = false;
    setTimeout(() => {
      const url = new URL(window.location.href);
      url.searchParams.set('tab', tab);
      window.history.replaceState({}, '', url.toString());
    }, 0);
  }

  onChallengesViewChanged(view: 'create' | 'manage'): void {
    this.challengesInitialView = view;
    setTimeout(() => {
      const url = new URL(window.location.href);
      url.searchParams.set('tab', view);
      window.history.replaceState({}, '', url.toString());
    }, 0);
  }

  showMessage(msg: string, type: 'success' | 'error'): void {
    this.message = msg;
    this.messageType = type;
    if (type === 'success') {
      setTimeout(() => {
        if (this.message === msg) this.message = '';
      }, 5000);
    }
  }

  toggleSidebar(): void {
    this.sidebarOpen = !this.sidebarOpen;
  }

  toggleMobileMenu(): void {
    this.mobileMenuOpen = !this.mobileMenuOpen;
  }
}
