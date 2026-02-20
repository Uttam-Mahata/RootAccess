import { Component, OnInit, OnDestroy, effect, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { ActivatedRoute, Router } from '@angular/router';
import Chart from 'chart.js/auto';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, FormArray } from '@angular/forms';
import { FormsModule } from '@angular/forms';
import { EditorModule, TINYMCE_SCRIPT_SRC } from '@tinymce/tinymce-angular';
import TurndownService from 'turndown';
import Showdown from 'showdown';
import { ChallengeService, ChallengeAdmin, ChallengeRequest, HintRequest } from '../../services/challenge';
import { NotificationService, Notification } from '../../services/notification';
import { ContestService } from '../../services/contest';
import { ContestAdminService, Contest, ContestRound } from '../../services/contest-admin';
import { AnalyticsService, AdminAnalytics } from '../../services/analytics';
import { AdminUserService, AdminUser } from '../../services/admin-user';
import { AdminTeamService, AdminTeam } from '../../services/admin-team';
import { BulkChallengeService } from '../../services/bulk-challenge';
import { ThemeService } from '../../services/theme';
import { ConfirmationModalService } from '../../services/confirmation-modal.service';
import { environment } from '../../../environments/environment';
import { take } from 'rxjs/operators';
import { DatetimePickerComponent } from '../datetime-picker/datetime-picker';

@Component({
  selector: 'app-admin-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    FormsModule,
    EditorModule,
    DatetimePickerComponent
  ],
  providers: [
    { provide: TINYMCE_SCRIPT_SRC, useValue: 'tinymce/tinymce.min.js' }
  ],
  templateUrl: './admin-dashboard.html',
  styleUrls: ['./admin-dashboard.scss']
})
export class AdminDashboardComponent implements OnInit, OnDestroy {
  challengeForm: FormGroup;
  notificationForm: FormGroup;
  contestForm: FormGroup;
  message = '';
  messageType: 'success' | 'error' = 'success';
  
  // Tab state
  activeTab: 'create' | 'manage' | 'notifications' | 'contest' | 'writeups' | 'audit' | 'analytics' | 'users' | 'teams' = 'create';
  
  // Sidebar state
  sidebarOpen = true;
  mobileMenuOpen = false;
  
  // Challenges list
  challenges: ChallengeAdmin[] = [];
  isLoading = false;
  
  // Notifications list
  notifications: Notification[] = [];
  isLoadingNotifications = false;
  isEditingNotification = false;
  editingNotificationId: string | null = null;
  
  // Edit mode
  isEditMode = false;
  editingChallengeId: string | null = null;
  
  // Preview mode
  previewChallenge: ChallengeAdmin | null = null;

  // Contest
  contestConfig: any = null;
  isLoadingContest = false;
  isLoadingContests = false;
  isLoadingRounds = false;
  contests: Contest[] = [];
  selectedContest: Contest | null = null;
  rounds: ContestRound[] = [];
  activeContestId: string | null = null;
  isCreatingContest = false;
  isEditingContest = false;
  editingContestId: string | null = null;
  contestEntityForm: FormGroup;
  isCreatingRound = false;
  isEditingRound = false;
  editingRoundId: string | null = null;
  roundForm: FormGroup;

  // Round-Challenge attachment
  selectedRound: ContestRound | null = null;
  roundChallengeIds: string[] = [];
  isLoadingRoundChallenges = false;
  availableChallengesForRound: ChallengeAdmin[] = [];
  selectedAttachedChallenges: Set<string> = new Set();
  selectedAvailableChallenges: Set<string> = new Set();

  // Official writeup
  officialWriteupChallengeId: string | null = null;
  officialWriteupContent = '';
  officialWriteupFormat: 'markdown' | 'html' = 'markdown';
  isSubmittingWriteup = false;

  // Writeups
  writeups: any[] = [];
  isLoadingWriteups = false;
  selectedTeamFilter: string = '';
  selectedWriteup: any | null = null;

  // Audit logs
  auditLogs: any[] = [];
  auditTotal = 0;
  auditPage = 1;
  isLoadingAudit = false;

  // Analytics
  analytics: AdminAnalytics | null = null;
  analyticsLoading = false;
  categoryKeys: string[] = [];
  difficultyKeys: string[] = [];
  private chartInstances: Chart[] = [];

  // User Management
  users: AdminUser[] = [];
  usersLoading = false;
  selectedUser: AdminUser | null = null;
  userScoreDelta: number = 0;
  userScoreReason: string = '';

  // Team Management
  teams: AdminTeam[] = [];
  teamsLoading = false;
  selectedTeam: AdminTeam | null = null;
  teamScoreDelta: number = 0;
  teamScoreReason: string = '';

  // Track which tabs have loaded their data to prevent duplicate loading
  private loadedTabs = new Set<string>();

  // Scoring types
  scoringTypes = [
    { value: 'dynamic', label: 'Dynamic (CTFd Formula)' },
    { value: 'linear', label: 'Linear Decay' },
    { value: 'static', label: 'Static (Fixed Points)' }
  ];
  
  // Rich text editor content (HTML)
  editorContent = '';
  
  // TinyMCE configuration
  editorConfig: any = {};
  editorKey = 0; // Used to force re-render when theme changes
  showEditor = true; // Used to control editor visibility during theme transitions
  
  // Predefined categories
  categories = [
    { value: 'web', label: 'Web Exploitation' },
    { value: 'crypto', label: 'Cryptography' },
    { value: 'pwn', label: 'Binary Exploitation (Pwn)' },
    { value: 'reverse', label: 'Reverse Engineering' },
    { value: 'forensics', label: 'Digital Forensics' },
    { value: 'networking', label: 'Networking' },
    { value: 'steganography', label: 'Steganography' },
    { value: 'osint', label: 'OSINT' },
    { value: 'misc', label: 'General Skills/Misc' }
  ];
  
  // Difficulty levels
  difficulties = [
    { value: 'easy', label: 'Easy', color: 'green' },
    { value: 'medium', label: 'Medium', color: 'yellow' },
    { value: 'hard', label: 'Hard', color: 'red' }
  ];

  // Notification types
  notificationTypes = [
    { value: 'info', label: 'Info', icon: 'info', colorClass: 'text-blue-500' },
    { value: 'warning', label: 'Warning', icon: 'warning', colorClass: 'text-amber-500' },
    { value: 'success', label: 'Success', icon: 'success', colorClass: 'text-emerald-500' },
    { value: 'error', label: 'Error', icon: 'error', colorClass: 'text-red-500' }
  ];

  // Markdown converter with enhanced configuration
  private turndownService = new TurndownService();
  private showdownConverter = new Showdown.Converter({
    tables: true,
    strikethrough: true,
    tasklists: true,
    smoothLivePreview: true,
    simpleLineBreaks: false,  // Proper paragraph handling
    openLinksInNewWindow: true,
    emoji: true,
    ghCodeBlocks: true,  // GitHub-style code blocks
    encodeEmails: true,
    simplifiedAutoLink: true,
    literalMidWordUnderscores: true,
    parseImgDimensions: true
  });

  constructor(
    private fb: FormBuilder,
    private challengeService: ChallengeService,
    private notificationService: NotificationService,
    private contestService: ContestService,
    private contestAdminService: ContestAdminService,
    private analyticsService: AnalyticsService,
    private adminUserService: AdminUserService,
    private adminTeamService: AdminTeamService,
    private bulkChallengeService: BulkChallengeService,
    private http: HttpClient,
    private themeService: ThemeService,
    private route: ActivatedRoute,
    private router: Router,
    private cdr: ChangeDetectorRef,
    private confirmationModalService: ConfirmationModalService
  ) {
    this.challengeForm = this.fb.group({
      title: ['', Validators.required],
      category: ['', Validators.required],
      difficulty: ['', Validators.required],
      description_format: ['markdown', Validators.required],
      max_points: [500, [Validators.required, Validators.min(1)]],
      min_points: [100, [Validators.required, Validators.min(1)]],
      decay: [10, [Validators.required, Validators.min(1)]],
      scoring_type: ['dynamic', Validators.required],
      flag: ['', Validators.required],
      files: [''],
      tags: ['']
    });

    this.notificationForm = this.fb.group({
      title: ['', Validators.required],
      content: ['', Validators.required],
      type: ['info', Validators.required]
    });

    this.contestForm = this.fb.group({
      title: ['', Validators.required],
      start_time: ['', Validators.required],
      end_time: ['', Validators.required],
      is_active: [false]
    });

    this.contestEntityForm = this.fb.group({
      name: ['', Validators.required],
      description: [''],
      start_time: ['', Validators.required],
      end_time: ['', Validators.required]
    });

    this.roundForm = this.fb.group({
      name: ['', Validators.required],
      description: [''],
      order: [0, Validators.required],
      visible_from: ['', Validators.required],
      start_time: ['', Validators.required],
      end_time: ['', Validators.required]
    });

    // Initialize editor config
    this.updateEditorConfig();

    // Watch for theme changes and update editor config
    effect(() => {
      this.themeService.isDarkMode();
      this.updateEditorConfig();
    });
  }

  private updateEditorConfig(): void {
    const isDark = this.themeService.isDarkMode();
    
    // Temporarily hide editor to force re-render with new theme
    this.showEditor = false;
    
    // Increment key to track theme changes
    this.editorKey++;
    
    this.editorConfig = {
      base_url: '/tinymce',
      suffix: '.min',
      height: 550,
      menubar: false,
      branding: false,
      promotion: false,
      plugins: [
        'advlist', 'autolink', 'lists', 'link', 'image', 'charmap',
        'anchor', 'searchreplace', 'visualblocks', 'code', 'codesample', 'fullscreen',
        'insertdatetime', 'media', 'table', 'preview', 'help', 'wordcount'
      ],
      toolbar: 'undo redo | blocks | bold italic forecolor backcolor | alignleft aligncenter alignright alignjustify | bullist numlist outdent indent | codesample code | removeformat | fullscreen | help',
      codesample_languages: [
        { text: 'HTML/XML', value: 'markup' },
        { text: 'JavaScript', value: 'javascript' },
        { text: 'TypeScript', value: 'typescript' },
        { text: 'CSS', value: 'css' },
        { text: 'Python', value: 'python' },
        { text: 'Java', value: 'java' },
        { text: 'C', value: 'c' },
        { text: 'C++', value: 'cpp' },
        { text: 'C#', value: 'csharp' },
        { text: 'PHP', value: 'php' },
        { text: 'Ruby', value: 'ruby' },
        { text: 'Go', value: 'go' },
        { text: 'Rust', value: 'rust' },
        { text: 'SQL', value: 'sql' },
        { text: 'Bash', value: 'bash' },
        { text: 'PowerShell', value: 'powershell' },
        { text: 'JSON', value: 'json' },
        { text: 'YAML', value: 'yaml' }
      ],
      content_style: isDark ? `
        body { 
          font-family: 'Space Grotesk', Arial, sans-serif; 
          font-size: 14px; 
          background-color: #1e293b;
          color: #e2e8f0;
          padding: 10px;
        }
        a { color: #f87171; text-decoration: underline; }
        code { 
          background-color: #0f172a; 
          padding: 3px 8px; 
          border-radius: 4px; 
          color: #fbbf24;
          font-family: 'Courier New', Courier, monospace;
          font-size: 13px;
        }
        pre { 
          background-color: #0f172a; 
          padding: 16px; 
          border-radius: 8px; 
          overflow-x: auto; 
          color: #e2e8f0;
          border: 1px solid #334155;
        }
        pre code { 
          background-color: transparent; 
          padding: 0; 
          color: #fbbf24; 
        }
      ` : `
        body { 
          font-family: 'Space Grotesk', Arial, sans-serif; 
          font-size: 14px; 
          background-color: #ffffff;
          color: #1e293b;
          padding: 10px;
        }
        a { color: #dc2626; text-decoration: underline; }
        code { 
          background-color: #f1f5f9; 
          padding: 3px 8px; 
          border-radius: 4px; 
          color: #b91c1c;
          font-family: 'Courier New', Courier, monospace;
          font-size: 13px;
        }
        pre { 
          background-color: #f1f5f9; 
          padding: 16px; 
          border-radius: 8px; 
          overflow-x: auto; 
          color: #1e293b;
          border: 1px solid #e2e8f0;
        }
        pre code { 
          background-color: transparent; 
          padding: 0; 
          color: #b91c1c; 
        }
      `,
      skin: isDark ? 'oxide-dark' : 'oxide',
      content_css: isDark ? 'dark' : 'default'
    };
    
    // Show editor again after a brief delay to ensure re-initialization
    setTimeout(() => {
      this.showEditor = true;
    }, 0);
  }

  ngOnInit(): void {
    // Always load initial data for sidebar counts
    this.loadChallenges();
    this.loadedTabs.add('manage'); // Mark as loaded
    this.loadTeams();
    this.loadedTabs.add('teams'); // Mark as loaded
    
    // Check for tab query parameter and set active tab (only on initial load)
    const params = this.route.snapshot.queryParams;
    const tab = params['tab'];
    if (tab && ['create', 'manage', 'notifications', 'contest', 'writeups', 'audit', 'analytics', 'users', 'teams'].includes(tab)) {
      this.activeTab = tab as typeof this.activeTab;
      // Only load data if not already loaded (manage and teams are already loaded above)
      if (!this.loadedTabs.has(tab)) {
        this.loadTabData(this.activeTab);
      }
    } else {
      // Default to create tab if no tab param
      this.activeTab = 'create';
    }
  }

  private loadTabData(tab: typeof this.activeTab): void {
    switch (tab) {
      case 'manage':
        if (!this.loadedTabs.has('manage')) {
          this.loadChallenges();
          this.loadedTabs.add('manage');
        }
        break;
      case 'notifications':
        if (!this.loadedTabs.has('notifications')) {
          this.loadNotifications();
          this.loadedTabs.add('notifications');
        }
        break;
      case 'contest':
        if (!this.loadedTabs.has('contest')) {
          this.loadContestConfig();
          this.loadContests();
          // Load challenges for round attachment UI
          if (this.challenges.length === 0) {
            this.loadChallenges();
          }
          this.loadedTabs.add('contest');
        }
        break;
      case 'writeups':
        if (!this.loadedTabs.has('writeups')) {
          // Load teams first for the filter dropdown
          if (this.teams.length === 0) {
            this.loadTeams();
          }
          this.loadWriteups();
          this.loadedTabs.add('writeups');
        }
        break;
      case 'audit':
        if (!this.loadedTabs.has('audit')) {
          this.loadAuditLogs();
          this.loadedTabs.add('audit');
        }
        break;
      case 'analytics':
        if (!this.loadedTabs.has('analytics')) {
          this.loadAnalytics();
          this.loadedTabs.add('analytics');
        }
        break;
      case 'users':
        if (!this.loadedTabs.has('users')) {
          this.loadUsers();
          this.loadedTabs.add('users');
        }
        break;
      case 'teams':
        if (!this.loadedTabs.has('teams')) {
          this.loadTeams();
          this.loadedTabs.add('teams');
        }
        break;
      case 'create':
        // No data to load for create tab
        break;
    }
  }

  loadChallenges(): void {
    this.isLoading = true;
    // Use list=1 for fast load (no descriptions); full data fetched on Preview/Edit
    this.challengeService.getChallengesForAdmin(true).subscribe({
      next: (data) => {
        this.challenges = data || [];
        this.isLoading = false;
      },
      error: (err) => {
        console.error('Error loading challenges:', err);
        this.challenges = [];
        this.isLoading = false;
        this.showMessage('Error loading challenges', 'error');
      }
    });
  }

  switchTab(tab: 'create' | 'manage' | 'notifications' | 'contest' | 'writeups' | 'audit' | 'analytics' | 'users' | 'teams'): void {
    // Prevent switching to the same tab
    if (this.activeTab === tab) {
      return;
    }
    
    // Set tab immediately for UI responsiveness
    this.activeTab = tab;
    this.mobileMenuOpen = false; // Close mobile menu on tab switch
    
    // Force change detection to ensure UI updates
    this.cdr.markForCheck();
    
    // Update URL query parameter asynchronously to avoid blocking UI update
    setTimeout(() => {
      const url = new URL(window.location.href);
      url.searchParams.set('tab', tab);
      window.history.replaceState({}, '', url.toString());
    }, 0);
    
    // Load data for the selected tab only if not already loaded
    this.loadTabData(tab);
    
    // For analytics tab, ensure charts initialize after DOM is ready
    if (tab === 'analytics' && this.analytics) {
      setTimeout(() => this.initCharts(), 400);
    }
    
    if (tab === 'create' && !this.isEditMode) {
      this.resetForm();
    }
  }

  toggleSidebar(): void {
    this.sidebarOpen = !this.sidebarOpen;
  }

  toggleMobileMenu(): void {
    this.mobileMenuOpen = !this.mobileMenuOpen;
  }

  // Notification methods
  loadNotifications(): void {
    this.isLoadingNotifications = true;
    this.notificationService.getAllNotifications().subscribe({
      next: (data) => {
        this.notifications = data || [];
        this.isLoadingNotifications = false;
      },
      error: (err) => {
        console.error('Error loading notifications:', err);
        this.notifications = [];
        this.isLoadingNotifications = false;
        this.showMessage('Error loading notifications', 'error');
      }
    });
  }

  onSubmitNotification(): void {
    if (this.notificationForm.valid) {
      const formValue = this.notificationForm.value;
      
      if (this.isEditingNotification && this.editingNotificationId) {
        // Update existing notification
        this.notificationService.updateNotification(this.editingNotificationId, {
          title: formValue.title,
          content: formValue.content,
          type: formValue.type,
          is_active: true
        }).subscribe({
          next: () => {
            this.showMessage('Notification updated successfully', 'success');
            this.loadNotifications();
            this.resetNotificationForm();
          },
          error: (err) => {
            console.error('Error updating notification:', err);
            this.showMessage('Error updating notification', 'error');
          }
        });
      } else {
        // Create new notification
        this.notificationService.createNotification({
          title: formValue.title,
          content: formValue.content,
          type: formValue.type
        }).subscribe({
          next: () => {
            this.showMessage('Notification created successfully', 'success');
            this.loadNotifications();
            this.resetNotificationForm();
          },
          error: (err) => {
            console.error('Error creating notification:', err);
            this.showMessage('Error creating notification', 'error');
          }
        });
      }
    }
  }

  editNotification(notification: Notification): void {
    this.isEditingNotification = true;
    this.editingNotificationId = notification.id;
    this.notificationForm.patchValue({
      title: notification.title,
      content: notification.content,
      type: notification.type
    });
  }

  deleteNotification(notification: Notification): void {
    this.confirmationModalService.show({
      title: 'Delete Notification',
      message: `Are you sure you want to delete the notification "${notification.title}"?`,
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        this.notificationService.deleteNotification(notification.id).subscribe({
          next: () => {
            this.showMessage('Notification deleted successfully', 'success');
            this.loadNotifications();
          },
          error: (err) => {
            console.error('Error deleting notification:', err);
            this.showMessage('Error deleting notification', 'error');
          }
        });
      }
    });
  }

  toggleNotificationActive(notification: Notification): void {
    this.notificationService.toggleNotificationActive(notification.id).subscribe({
      next: () => {
        this.showMessage(`Notification ${notification.is_active ? 'deactivated' : 'activated'} successfully`, 'success');
        this.loadNotifications();
      },
      error: (err) => {
        console.error('Error toggling notification:', err);
        this.showMessage('Error toggling notification status', 'error');
      }
    });
  }

  resetNotificationForm(): void {
    this.isEditingNotification = false;
    this.editingNotificationId = null;
    this.notificationForm.reset({
      title: '',
      content: '',
      type: 'info'
    });
  }

  cancelEditNotification(): void {
    this.resetNotificationForm();
  }

  getNotificationTypeLabel(value: string): string {
    const type = this.notificationTypes.find(t => t.value === value);
    return type ? type.label : value;
  }

  getNotificationTypeColorClass(value: string): string {
    const type = this.notificationTypes.find(t => t.value === value);
    return type ? type.colorClass : 'text-slate-500';
  }

  onEditorChange(event: any): void {
    this.editorContent = event.editor.getContent();
  }

  onSubmit(): void {
    if (this.challengeForm.valid && this.editorContent.trim()) {
      const formValue = this.challengeForm.value;
      const selectedFormat = formValue.description_format || 'markdown';
      
      // Convert HTML to selected format
      let description: string;
      if (selectedFormat === 'markdown') {
        // Convert TinyMCE HTML to Markdown
        description = this.turndownService.turndown(this.editorContent);
      } else {
        // Store as HTML directly
        description = this.editorContent;
      }
      
      const challenge: ChallengeRequest = {
        title: formValue.title,
        description: description,
        description_format: selectedFormat,
        category: formValue.category,
        difficulty: formValue.difficulty,
        max_points: formValue.max_points,
        min_points: formValue.min_points,
        decay: formValue.decay,
        scoring_type: formValue.scoring_type || 'dynamic',
        flag: formValue.flag,
        files: formValue.files ? formValue.files.split(',').map((f: string) => f.trim()).filter((f: string) => f) : [],
        tags: formValue.tags ? formValue.tags.split(',').map((t: string) => t.trim()).filter((t: string) => t) : [],
        hints: []
      };

      if (this.isEditMode && this.editingChallengeId) {
        // Update existing challenge
        this.challengeService.updateChallenge(this.editingChallengeId, challenge).subscribe({
          next: () => {
            this.showMessage('Challenge updated successfully', 'success');
            // Optimized: Only reload challenges if on manage tab
            if (this.activeTab === 'manage') {
              this.loadChallenges();
            }
            this.resetForm();
            this.switchTab('manage');
          },
          error: (err) => {
            console.error('Error updating challenge:', err);
            this.showMessage('Error updating challenge', 'error');
          }
        });
      } else {
        // Create new challenge
        this.challengeService.createChallenge(challenge).subscribe({
          next: () => {
            this.showMessage('Challenge created successfully', 'success');
            this.loadChallenges();
            this.resetForm();
          },
          error: (err) => {
            console.error('Error creating challenge:', err);
            this.showMessage('Error creating challenge', 'error');
          }
        });
      }
    } else if (!this.editorContent.trim()) {
      this.showMessage('Please provide a description for the challenge', 'error');
    }
  }

  previewChallengeToggle(challenge: ChallengeAdmin): void {
    // List API omits description; fetch full challenge for preview
    if (!challenge.description) {
      this.challengeService.getChallenge(challenge.id).subscribe({
        next: (full) => {
          this.previewChallenge = { ...challenge, description: full.description, description_format: full.description_format };
        },
        error: () => this.showMessage('Failed to load challenge details', 'error')
      });
    } else {
      this.previewChallenge = challenge;
    }
  }

  editChallenge(challenge: ChallengeAdmin): void {
    this.isEditMode = true;
    this.editingChallengeId = challenge.id;
    
    const applyEdit = (ch: ChallengeAdmin) => {
      const format = ch.description_format || 'markdown';
      if (format === 'html') {
        this.editorContent = ch.description;
      } else {
        this.editorContent = this.showdownConverter.makeHtml(ch.description || '');
      }
      // Remove required validator on flag for edit mode
      this.challengeForm.get('flag')?.clearValidators();
      this.challengeForm.get('flag')?.updateValueAndValidity();
      this.challengeForm.patchValue({
        title: ch.title,
        category: ch.category,
        difficulty: ch.difficulty,
        description_format: format,
        max_points: ch.max_points,
        min_points: ch.min_points,
        decay: ch.decay,
        scoring_type: ch.scoring_type || 'dynamic',
        flag: '',
        files: ch.files ? ch.files.join(', ') : '',
        tags: ch.tags ? ch.tags.join(', ') : ''
      });
      this.switchTab('create');
      this.showMessage(`Editing: ${ch.title} (Leave flag empty to keep current flag)`, 'success');
    };

    // List API omits description; fetch full challenge for edit
    if (!challenge.description) {
      this.challengeService.getChallenge(challenge.id).subscribe({
        next: (full) => {
          applyEdit({ ...challenge, description: full.description, description_format: full.description_format });
        },
        error: () => this.showMessage('Failed to load challenge for edit', 'error')
      });
    } else {
      applyEdit(challenge);
    }
  }

  deleteChallenge(challenge: ChallengeAdmin): void {
    this.confirmationModalService.show({
      title: 'Delete Challenge',
      message: `Are you sure you want to delete "${challenge.title}"? This action cannot be undone.`,
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        this.challengeService.deleteChallenge(challenge.id).subscribe({
          next: () => {
            this.showMessage('Challenge deleted successfully', 'success');
            this.loadChallenges();
          },
          error: (err) => {
            console.error('Error deleting challenge:', err);
            this.showMessage('Error deleting challenge', 'error');
          }
        });
      }
    });
  }

  resetForm(): void {
    this.isEditMode = false;
    this.editingChallengeId = null;
    this.editorContent = '';
    // Restore required validator for flag (needed for create mode)
    this.challengeForm.get('flag')?.setValidators(Validators.required);
    this.challengeForm.get('flag')?.updateValueAndValidity();
    this.challengeForm.reset({
      title: '',
      category: '',
      difficulty: '',
      description_format: 'markdown',
      max_points: 500,
      min_points: 100,
      decay: 10,
      scoring_type: 'dynamic',
      flag: '',
      files: '',
      tags: ''
    });
    this.message = '';
  }

  cancelEdit(): void {
    this.resetForm();
    this.switchTab('manage');
  }

  getCategoryLabel(value: string): string {
    const category = this.categories.find(c => c.value === value);
    return category ? category.label : value;
  }

  getDifficultyLabel(value: string): string {
    const difficulty = this.difficulties.find(d => d.value === value);
    return difficulty ? difficulty.label : value;
  }

  getDifficultyColor(value: string): string {
    const difficulty = this.difficulties.find(d => d.value === value);
    return difficulty ? difficulty.color : 'gray';
  }

  private showMessage(msg: string, type: 'success' | 'error'): void {
    this.message = msg;
    this.messageType = type;
    
    // Auto-clear success messages after 5 seconds
    if (type === 'success') {
      setTimeout(() => {
        if (this.message === msg) {
          this.message = '';
        }
      }, 5000);
    }
  }

  // Contest management
  loadContestConfig(): void {
    this.isLoadingContest = true;
    this.contestService.getContestConfig().subscribe({
      next: (data) => {
        this.contestConfig = data.config || null;
        this.activeContestId = this.contestConfig?.contest_id || null;
        if (this.contestConfig) {
          this.contestForm.patchValue({
            title: this.contestConfig.title,
            start_time: this.formatDateForInput(this.contestConfig.start_time),
            end_time: this.formatDateForInput(this.contestConfig.end_time),
            is_active: this.contestConfig.is_active
          });
        }
        this.isLoadingContest = false;
      },
      error: () => {
        this.contestConfig = null;
        this.activeContestId = null;
        this.isLoadingContest = false;
      }
    });
  }

  loadContests(): void {
    this.isLoadingContests = true;
    this.contestAdminService.listContests().subscribe({
      next: (data) => {
        this.contests = data || [];
        this.isLoadingContests = false;
      },
      error: () => {
        this.contests = [];
        this.isLoadingContests = false;
      }
    });
  }

  loadRounds(): void {
    if (!this.selectedContest?.id) return;
    this.isLoadingRounds = true;
    this.contestAdminService.listRounds(this.selectedContest.id).subscribe({
      next: (data) => {
        this.rounds = (data || []).sort((a, b) => a.order - b.order);
        this.isLoadingRounds = false;
      },
      error: () => {
        this.rounds = [];
        this.isLoadingRounds = false;
      }
    });
  }

  selectContest(contest: Contest): void {
    this.selectedContest = contest;
    this.loadRounds();
    this.isEditingContest = false;
    this.editingContestId = null;
    this.isCreatingRound = false;
    this.isEditingRound = false;
    this.editingRoundId = null;
    this.selectedRound = null;
    this.roundChallengeIds = [];
    this.availableChallengesForRound = [];
  }

  startCreateContest(): void {
    this.isCreatingContest = true;
    this.isEditingContest = false;
    this.editingContestId = null;
    this.contestEntityForm.reset({ name: '', description: '', start_time: '', end_time: '' });
  }

  startEditContest(contest: Contest): void {
    this.isCreatingContest = false;
    this.isEditingContest = true;
    this.editingContestId = contest.id;
    this.contestEntityForm.patchValue({
      name: contest.name,
      description: contest.description || '',
      start_time: contest.start_time ? new Date(contest.start_time).toISOString() : '',
      end_time: contest.end_time ? new Date(contest.end_time).toISOString() : ''
    });
  }

  cancelContestForm(): void {
    this.isCreatingContest = false;
    this.isEditingContest = false;
    this.editingContestId = null;
    this.contestEntityForm.reset();
  }

  onSubmitContestEntity(): void {
    if (!this.contestEntityForm.valid) return;
    const v = this.contestEntityForm.value;
    const startTime = new Date(v.start_time).toISOString();
    const endTime = new Date(v.end_time).toISOString();
    if (this.isEditingContest && this.editingContestId) {
      // Optimistic update
      const contestIndex = this.contests.findIndex(c => c.id === this.editingContestId);
      let originalContest: Contest | null = null;
      if (contestIndex !== -1) {
        originalContest = { ...this.contests[contestIndex] };
        this.contests[contestIndex] = {
          ...this.contests[contestIndex],
          name: v.name,
          description: v.description || '',
          start_time: startTime,
          end_time: endTime
        };
        if (this.selectedContest?.id === this.editingContestId) {
          this.selectedContest = { ...this.contests[contestIndex] };
        }
      }
      this.cancelContestForm();
      this.isCreatingContest = false;
      
      this.contestAdminService.updateContest(this.editingContestId, v.name, v.description || '', startTime, endTime, false).subscribe({
        next: (updated) => {
          this.showMessage('Contest updated', 'success');
          // Update contest in list with server response
          const index = this.contests.findIndex(c => c.id === updated.id);
          if (index !== -1) {
            this.contests[index] = updated;
          } else {
            // Fallback: reload if contest not found
            this.loadContests();
          }
          if (this.selectedContest?.id === this.editingContestId) {
            this.selectedContest = updated;
            // Only reload rounds if contest was selected
            if (this.selectedRound) {
              this.loadRounds();
            }
          }
        },
        error: (err) => {
          // Revert optimistic update
          if (contestIndex !== -1 && originalContest) {
            this.contests[contestIndex] = originalContest;
            if (this.selectedContest?.id === this.editingContestId) {
              this.selectedContest = { ...originalContest };
            }
          }
          this.showMessage(err.error?.error || 'Error updating contest', 'error');
        }
      });
    } else {
      // Optimistic update - add temporary contest
      const tempContest: Contest = {
        id: 'temp-' + Date.now(),
        name: v.name,
        description: v.description || '',
        start_time: startTime,
        end_time: endTime,
        is_active: false
      };
      this.contests = [tempContest, ...this.contests];
      this.cancelContestForm();
      this.isCreatingContest = false;
      
      this.contestAdminService.createContest(v.name, v.description || '', startTime, endTime).subscribe({
        next: (created) => {
          this.showMessage('Contest created', 'success');
          // Replace temp with real contest
          const tempIndex = this.contests.findIndex(c => c.id === tempContest.id);
          if (tempIndex !== -1) {
            this.contests[tempIndex] = created;
          } else {
            // Fallback: reload if temp contest not found
            this.loadContests();
          }
        },
        error: (err) => {
          // Remove temp contest
          this.contests = this.contests.filter(c => c.id !== tempContest.id);
          this.showMessage(err.error?.error || 'Error creating contest', 'error');
        }
      });
    }
  }

  setActiveContest(contestId: string): void {
    this.contestAdminService.setActiveContest(contestId).subscribe({
      next: () => {
        this.showMessage('Active contest updated', 'success');
        this.activeContestId = contestId;
        this.loadContestConfig();
      },
      error: (err) => this.showMessage(err.error?.error || 'Error setting active contest', 'error')
    });
  }

  deleteContest(contest: Contest): void {
    this.confirmationModalService.show({
      title: 'Delete Contest',
      message: `Are you sure you want to delete "${contest.name}"? This will also delete all rounds.`,
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (!confirmed) return;
      
      // Optimistic update
      const wasSelected = this.selectedContest?.id === contest.id;
      const wasActive = this.activeContestId === contest.id;
      const originalContest = { ...contest }; // Store for error revert
      this.contests = this.contests.filter(c => c.id !== contest.id);
      if (wasSelected) {
        this.selectedContest = null;
        this.rounds = [];
      }
      if (wasActive) {
        this.activeContestId = null;
      }
      
      this.contestAdminService.deleteContest(contest.id).subscribe({
        next: () => {
          this.showMessage('Contest deleted', 'success');
          // Refresh to sync
          this.loadContests();
          if (wasActive) {
            this.loadContestConfig();
          }
        },
        error: (err) => {
          // Revert optimistic update - restore contest to list
          this.contests = [...this.contests, originalContest];
          if (wasSelected) {
            this.selectContest(originalContest);
          }
          if (wasActive) {
            this.activeContestId = contest.id;
            this.loadContestConfig();
          }
          // Refresh to sync with server
          this.loadContests();
          this.showMessage(err.error?.error || 'Error deleting contest', 'error');
        }
      });
    });
  }

  getContestStatus(contest: Contest): 'not_started' | 'running' | 'ended' {
    const now = new Date().getTime();
    const start = new Date(contest.start_time).getTime();
    const end = new Date(contest.end_time).getTime();
    if (now < start) return 'not_started';
    if (now > end) return 'ended';
    return 'running';
  }

  startCreateRound(): void {
    this.isCreatingRound = true;
    this.isEditingRound = false;
    this.editingRoundId = null;
    this.roundForm.reset({ name: '', description: '', order: this.rounds.length, visible_from: '', start_time: '', end_time: '' });
  }

  startEditRound(round: ContestRound): void {
    this.isCreatingRound = false;
    this.isEditingRound = true;
    this.editingRoundId = round.id;
    this.roundForm.patchValue({
      name: round.name,
      description: round.description || '',
      order: round.order,
      visible_from: round.visible_from ? new Date(round.visible_from).toISOString() : '',
      start_time: round.start_time ? new Date(round.start_time).toISOString() : '',
      end_time: round.end_time ? new Date(round.end_time).toISOString() : ''
    });
  }

  cancelRoundForm(): void {
    this.isCreatingRound = false;
    this.isEditingRound = false;
    this.editingRoundId = null;
    this.roundForm.reset();
  }

  onSubmitRound(): void {
    if (!this.roundForm.valid || !this.selectedContest) return;
    const v = this.roundForm.value;
    const visibleFrom = new Date(v.visible_from).toISOString();
    const startTime = new Date(v.start_time).toISOString();
    const endTime = new Date(v.end_time).toISOString();
    if (this.isEditingRound && this.editingRoundId) {
      // Optimistic update
      const roundIndex = this.rounds.findIndex(r => r.id === this.editingRoundId);
      let originalRound: ContestRound | null = null;
      if (roundIndex !== -1) {
        originalRound = { ...this.rounds[roundIndex] };
        this.rounds[roundIndex] = {
          ...this.rounds[roundIndex],
          name: v.name,
          description: v.description || '',
          order: v.order,
          visible_from: visibleFrom,
          start_time: startTime,
          end_time: endTime
        };
        this.rounds.sort((a, b) => a.order - b.order);
        if (this.selectedRound?.id === this.editingRoundId) {
          const newIndex = this.rounds.findIndex(r => r.id === this.editingRoundId);
          if (newIndex !== -1) {
            this.selectedRound = { ...this.rounds[newIndex] };
          }
        }
      }
      this.cancelRoundForm();
      this.isCreatingRound = false;
      
      this.contestAdminService.updateRound(this.selectedContest.id, this.editingRoundId, v.name, v.description || '', v.order, visibleFrom, startTime, endTime).subscribe({
        next: (updated) => {
          this.showMessage('Round updated', 'success');
          // Update the round in the list with server response
          const index = this.rounds.findIndex(r => r.id === updated.id);
          if (index !== -1) {
            this.rounds[index] = updated;
            this.rounds.sort((a, b) => a.order - b.order);
          } else {
            // Fallback: reload if round not found
            this.loadRounds();
          }
        },
        error: (err) => {
          // Revert optimistic update
          if (roundIndex !== -1 && originalRound) {
            this.rounds[roundIndex] = originalRound;
            this.rounds.sort((a, b) => a.order - b.order);
            if (this.selectedRound?.id === this.editingRoundId) {
              this.selectedRound = { ...originalRound };
            }
          }
          this.showMessage(err.error?.error || 'Error updating round', 'error');
        }
      });
    } else {
      // Optimistic update - add temporary round
      const tempRound: ContestRound = {
        id: 'temp-' + Date.now(),
        contest_id: this.selectedContest.id,
        name: v.name,
        description: v.description || '',
        order: v.order,
        visible_from: visibleFrom,
        start_time: startTime,
        end_time: endTime
      };
      this.rounds = [...this.rounds, tempRound].sort((a, b) => a.order - b.order);
      this.cancelRoundForm();
      this.isCreatingRound = false;
      
      this.contestAdminService.createRound(this.selectedContest.id, v.name, v.description || '', v.order, visibleFrom, startTime, endTime).subscribe({
        next: (created) => {
          this.showMessage('Round created', 'success');
          // Replace temp with real round
          const tempIndex = this.rounds.findIndex(r => r.id === tempRound.id);
          if (tempIndex !== -1) {
            this.rounds[tempIndex] = created;
            this.rounds.sort((a, b) => a.order - b.order);
          } else {
            // Fallback: reload if temp round not found
            this.loadRounds();
          }
        },
        error: (err) => {
          // Remove temp round
          this.rounds = this.rounds.filter(r => r.id !== tempRound.id);
          this.showMessage(err.error?.error || 'Error creating round', 'error');
        }
      });
    }
  }

  deleteRound(round: ContestRound): void {
    if (!this.selectedContest) return;
    this.confirmationModalService.show({
      title: 'Delete Round',
      message: `Are you sure you want to delete "${round.name}"?`,
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (!confirmed) return;
      
      // Optimistic update
      const wasSelected = this.selectedRound?.id === round.id;
      const originalRounds = [...this.rounds];
      this.rounds = this.rounds.filter(r => r.id !== round.id);
      if (wasSelected) {
        this.selectedRound = null;
        this.roundChallengeIds = [];
        this.availableChallengesForRound = [];
      }
      this.cancelRoundForm();
      
      this.contestAdminService.deleteRound(this.selectedContest!.id, round.id).subscribe({
        next: () => {
          this.showMessage('Round deleted', 'success');
          // Refresh to sync
          this.loadRounds();
        },
        error: (err) => {
          // Revert optimistic update
          this.rounds = originalRounds;
          if (wasSelected) {
            this.selectedRound = round;
            this.loadRoundChallenges();
          }
          this.showMessage(err.error?.error || 'Error deleting round', 'error');
        }
      });
    });
  }

  getRoundStatus(round: ContestRound): 'not_started' | 'running' | 'ended' {
    const now = new Date().getTime();
    const start = new Date(round.start_time).getTime();
    const end = new Date(round.end_time).getTime();
    if (now < start) return 'not_started';
    if (now > end) return 'ended';
    return 'running';
  }

  onSubmitContest(): void {
    if (this.contestForm.valid) {
      const formValue = this.contestForm.value;
      const startTime = new Date(formValue.start_time).toISOString();
      const endTime = new Date(formValue.end_time).toISOString();

      this.contestService.updateContestConfig(
        formValue.title,
        startTime,
        endTime,
        formValue.is_active
      ).subscribe({
        next: () => {
          this.showMessage('Contest configuration updated', 'success');
          this.loadContestConfig();
        },
        error: (err) => {
          this.showMessage(err.error?.error || 'Error updating contest config', 'error');
        }
      });
    }
  }

  formatDateForInput(dateStr: string): string {
    if (!dateStr) return '';
    const d = new Date(dateStr);
    return d.toISOString().slice(0, 16);
  }

  formatDateForDisplay(dateStr: string): string {
    if (!dateStr) return '';
    const d = new Date(dateStr);
    return d.toLocaleString('en-US', {
      month: 'short', day: 'numeric', year: 'numeric',
      hour: '2-digit', minute: '2-digit'
    });
  }

  // Round-Challenge attachment
  selectRound(round: ContestRound): void {
    this.selectedRound = this.selectedRound?.id === round.id ? null : round;
    if (this.selectedRound) {
      // Clear selections when switching rounds
      this.selectedAttachedChallenges.clear();
      this.selectedAvailableChallenges.clear();
      // Ensure challenges are loaded for attachment UI
      if (this.challenges.length === 0 && this.activeTab === 'contest') {
        this.loadChallenges();
      }
      this.loadRoundChallenges();
    } else {
      this.roundChallengeIds = [];
      this.availableChallengesForRound = [];
      this.selectedAttachedChallenges.clear();
      this.selectedAvailableChallenges.clear();
    }
  }

  loadRoundChallenges(): void {
    if (!this.selectedContest || !this.selectedRound) return;
    this.isLoadingRoundChallenges = true;
    this.contestAdminService.getRoundChallenges(
      this.selectedContest.id, this.selectedRound.id
    ).subscribe({
      next: (ids) => {
        this.roundChallengeIds = ids || [];
        this.isLoadingRoundChallenges = false;
        this.computeAvailableChallenges();
        // Clear selections after loading
        this.selectedAttachedChallenges.clear();
        this.selectedAvailableChallenges.clear();
      },
      error: () => {
        this.roundChallengeIds = [];
        this.isLoadingRoundChallenges = false;
      }
    });
  }

  computeAvailableChallenges(): void {
    const attachedSet = new Set(this.roundChallengeIds);
    this.availableChallengesForRound = this.challenges.filter(
      ch => !attachedSet.has(ch.id)
    );
  }

  attachChallenge(challengeId: string): void {
    if (!this.selectedContest || !this.selectedRound) return;
    this.attachChallenges([challengeId]);
  }

  attachChallenges(challengeIds: string[]): void {
    if (!this.selectedContest || !this.selectedRound || challengeIds.length === 0) return;
    
    // Optimistic update
    const newIds = challengeIds.filter(id => !this.roundChallengeIds.includes(id));
    this.roundChallengeIds = [...this.roundChallengeIds, ...newIds];
    this.computeAvailableChallenges();
    this.selectedAvailableChallenges.clear();
    
    this.contestAdminService.attachChallenges(
      this.selectedContest.id, this.selectedRound.id, challengeIds
    ).subscribe({
      next: () => {
        this.showMessage(
          challengeIds.length === 1 
            ? 'Challenge attached to round' 
            : `${challengeIds.length} challenges attached to round`, 
          'success'
        );
        // Refresh to sync
        this.loadRoundChallenges();
      },
      error: (err) => {
        // Revert optimistic update
        this.roundChallengeIds = this.roundChallengeIds.filter(id => !newIds.includes(id));
        this.computeAvailableChallenges();
        this.showMessage(
          err.error?.error || 'Error attaching challenges', 'error'
        );
      }
    });
  }

  detachChallenge(challengeId: string): void {
    if (!this.selectedContest || !this.selectedRound) return;
    this.detachChallenges([challengeId]);
  }

  detachChallenges(challengeIds: string[]): void {
    if (!this.selectedContest || !this.selectedRound || challengeIds.length === 0) return;
    
    this.confirmationModalService.show({
      title: challengeIds.length === 1 ? 'Detach Challenge' : 'Detach Challenges',
      message: challengeIds.length === 1 
        ? 'Remove this challenge from the round?'
        : `Remove ${challengeIds.length} challenges from the round?`,
      confirmText: 'Remove',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (!confirmed) return;
      
      // Optimistic update
      const hadChallenges = challengeIds.filter(id => this.roundChallengeIds.includes(id));
      this.roundChallengeIds = this.roundChallengeIds.filter(id => !challengeIds.includes(id));
      this.computeAvailableChallenges();
      this.selectedAttachedChallenges.clear();
      
      this.contestAdminService.detachChallenges(
        this.selectedContest!.id, this.selectedRound!.id, challengeIds
      ).subscribe({
        next: () => {
          this.showMessage(
            challengeIds.length === 1 
              ? 'Challenge detached from round' 
              : `${challengeIds.length} challenges detached from round`, 
            'success'
          );
          // Refresh to sync
          this.loadRoundChallenges();
        },
        error: (err) => {
          // Revert optimistic update
          this.roundChallengeIds = [...this.roundChallengeIds, ...hadChallenges];
          this.computeAvailableChallenges();
          this.showMessage(
            err.error?.error || 'Error detaching challenges', 'error'
          );
        }
      });
    });
  }

  // Checkbox selection helpers
  toggleAttachedChallenge(challengeId: string): void {
    if (this.selectedAttachedChallenges.has(challengeId)) {
      this.selectedAttachedChallenges.delete(challengeId);
    } else {
      this.selectedAttachedChallenges.add(challengeId);
    }
  }

  toggleAvailableChallenge(challengeId: string): void {
    if (this.selectedAvailableChallenges.has(challengeId)) {
      this.selectedAvailableChallenges.delete(challengeId);
    } else {
      this.selectedAvailableChallenges.add(challengeId);
    }
  }

  isAttachedChallengeSelected(challengeId: string): boolean {
    return this.selectedAttachedChallenges.has(challengeId);
  }

  isAvailableChallengeSelected(challengeId: string): boolean {
    return this.selectedAvailableChallenges.has(challengeId);
  }

  selectAllAttached(): void {
    this.roundChallengeIds.forEach(id => this.selectedAttachedChallenges.add(id));
  }

  deselectAllAttached(): void {
    this.selectedAttachedChallenges.clear();
  }

  selectAllAvailable(): void {
    this.availableChallengesForRound.forEach(ch => this.selectedAvailableChallenges.add(ch.id));
  }

  deselectAllAvailable(): void {
    this.selectedAvailableChallenges.clear();
  }

  areAllAttachedSelected(): boolean {
    return this.roundChallengeIds.length > 0 && 
           this.roundChallengeIds.every(id => this.selectedAttachedChallenges.has(id));
  }

  areAllAvailableSelected(): boolean {
    return this.availableChallengesForRound.length > 0 && 
           this.availableChallengesForRound.every(ch => this.selectedAvailableChallenges.has(ch.id));
  }

  bulkAttachSelected(): void {
    const selectedIds = Array.from(this.selectedAvailableChallenges);
    if (selectedIds.length > 0) {
      this.attachChallenges(selectedIds);
    }
  }

  bulkDetachSelected(): void {
    const selectedIds = Array.from(this.selectedAttachedChallenges);
    if (selectedIds.length > 0) {
      this.detachChallenges(selectedIds);
    }
  }

  getChallengeTitle(challengeId: string): string {
    const ch = this.challenges.find(c => c.id === challengeId);
    return ch ? ch.title : challengeId;
  }

  getChallengeCategory(challengeId: string): string {
    const ch = this.challenges.find(c => c.id === challengeId);
    return ch ? this.getCategoryLabel(ch.category) : '';
  }

  getChallengeDifficulty(challengeId: string): string {
    const ch = this.challenges.find(c => c.id === challengeId);
    return ch ? ch.difficulty : '';
  }

  getChallengePoints(challengeId: string): number {
    const ch = this.challenges.find(c => c.id === challengeId);
    return ch ? ch.current_points : 0;
  }

  // Official writeup management
  startEditOfficialWriteup(challenge: ChallengeAdmin): void {
    this.officialWriteupChallengeId = challenge.id;
    this.officialWriteupContent = (challenge as any).official_writeup || '';
    this.officialWriteupFormat = (challenge as any).official_writeup_format || 'markdown';
  }

  cancelOfficialWriteup(): void {
    this.officialWriteupChallengeId = null;
    this.officialWriteupContent = '';
  }

  saveOfficialWriteup(): void {
    if (!this.officialWriteupChallengeId) return;
    this.isSubmittingWriteup = true;
    this.challengeService.updateOfficialWriteup(
      this.officialWriteupChallengeId, this.officialWriteupContent, this.officialWriteupFormat
    ).subscribe({
      next: () => {
        this.showMessage('Official writeup saved', 'success');
        this.isSubmittingWriteup = false;
      },
      error: () => {
        this.showMessage('Error saving writeup', 'error');
        this.isSubmittingWriteup = false;
      }
    });
  }

  publishOfficialWriteup(challengeId: string): void {
    this.confirmationModalService.show({
      title: 'Publish Official Writeup',
      message: 'Publish this writeup? It will be visible to users after the contest ends.',
      confirmText: 'Publish',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (!confirmed) return;
      this.challengeService.publishOfficialWriteup(challengeId).subscribe({
        next: () => this.showMessage('Official writeup published', 'success'),
        error: (err) => this.showMessage(err.error?.error || 'Cannot publish yet (contest may not have ended)', 'error')
      });
    });
  }

  getScoringTypeLabel(value: string): string {
    const type = this.scoringTypes.find(t => t.value === value);
    return type ? type.label : value || 'Dynamic';
  }

  // Writeup management
  loadWriteups(): void {
    this.isLoadingWriteups = true;
    const url = this.selectedTeamFilter 
      ? `${this.apiUrl}/admin/writeups?team_id=${this.selectedTeamFilter}`
      : `${this.apiUrl}/admin/writeups`;
    this.http.get<any[]>(url).subscribe({
      next: (data) => {
        this.writeups = data || [];
        this.isLoadingWriteups = false;
      },
      error: () => {
        this.writeups = [];
        this.isLoadingWriteups = false;
      }
    });
  }

  onTeamFilterChange(): void {
    this.loadWriteups();
  }

  approveWriteup(id: string): void {
    // Optimistic update
    const writeup = this.writeups.find(w => w.id === id);
    if (writeup) {
      const originalStatus = writeup.status;
      writeup.status = 'approved';
      
      this.http.put<any>(`${this.apiUrl}/admin/writeups/${id}/status`, { status: 'approved' }).subscribe({
        next: () => {
          this.showMessage('Writeup approved', 'success');
        },
        error: () => {
          // Revert on error
          writeup.status = originalStatus;
          this.showMessage('Error approving writeup', 'error');
        }
      });
    }
  }

  rejectWriteup(id: string): void {
    // Optimistic update
    const writeup = this.writeups.find(w => w.id === id);
    if (writeup) {
      const originalStatus = writeup.status;
      writeup.status = 'rejected';
      
      this.http.put<any>(`${this.apiUrl}/admin/writeups/${id}/status`, { status: 'rejected' }).subscribe({
        next: () => {
          this.showMessage('Writeup rejected', 'success');
        },
        error: () => {
          // Revert on error
          writeup.status = originalStatus;
          this.showMessage('Error rejecting writeup', 'error');
        }
      });
    }
  }

  deleteWriteup(id: string): void {
    this.confirmationModalService.show({
      title: 'Delete Writeup',
      message: 'Are you sure you want to delete this writeup?',
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        // Optimistic update
        const index = this.writeups.findIndex(w => w.id === id);
        if (index !== -1) {
          const writeup = this.writeups[index];
          this.writeups.splice(index, 1);
          
          // Close modal if viewing this writeup
          if (this.selectedWriteup?.id === id) {
            this.selectedWriteup = null;
          }
          
          this.http.delete<any>(`${this.apiUrl}/admin/writeups/${id}`).subscribe({
            next: () => {
              this.showMessage('Writeup deleted', 'success');
            },
            error: () => {
              // Revert on error
              this.writeups.splice(index, 0, writeup);
              this.showMessage('Error deleting writeup', 'error');
            }
          });
        }
      }
    });
  }

  viewWriteup(writeup: any): void {
    this.selectedWriteup = writeup;
  }

  // Audit log
  loadAuditLogs(): void {
    this.isLoadingAudit = true;
    this.http.get<any>(`${this.apiUrl}/admin/audit-logs?page=${this.auditPage}&limit=50`).subscribe({
      next: (data) => {
        this.auditLogs = data.logs || [];
        this.auditTotal = data.total || 0;
        this.isLoadingAudit = false;
      },
      error: () => {
        this.auditLogs = [];
        this.isLoadingAudit = false;
      }
    });
  }

  nextAuditPage(): void {
    this.auditPage++;
    this.loadAuditLogs();
  }

  prevAuditPage(): void {
    if (this.auditPage > 1) {
      this.auditPage--;
      this.loadAuditLogs();
    }
  }

  // Analytics
  loadAnalytics(): void {
    this.analyticsLoading = true;
    this.destroyCharts();
    this.analyticsService.getPlatformAnalytics().subscribe({
      next: (data) => {
        this.analytics = data;
        this.categoryKeys = data.category_breakdown ? Object.keys(data.category_breakdown) : [];
        this.difficultyKeys = data.difficulty_breakdown ? Object.keys(data.difficulty_breakdown) : [];
        this.analyticsLoading = false;
        // Wait for Angular to render the canvas elements before initializing charts
        // Use requestAnimationFrame to ensure DOM is ready
        requestAnimationFrame(() => {
          setTimeout(() => this.initCharts(), 100);
        });
      },
      error: () => {
        this.analyticsLoading = false;
        this.showMessage('Failed to load analytics', 'error');
      }
    });
  }

  private destroyCharts(): void {
    this.chartInstances.forEach(chart => chart.destroy());
    this.chartInstances = [];
  }

  private initCharts(): void {
    if (!this.analytics || this.activeTab !== 'analytics') return;
    
    // Ensure canvas elements exist before initializing
    const solvesEl = document.getElementById('solvesOverTimeChart') as HTMLCanvasElement;
    if (!solvesEl) {
      // Canvas not ready yet, retry after a short delay (max 5 retries)
      const retryCount = (this as any)._chartRetryCount || 0;
      if (retryCount < 5) {
        (this as any)._chartRetryCount = retryCount + 1;
        setTimeout(() => this.initCharts(), 150);
      } else {
        (this as any)._chartRetryCount = 0;
      }
      return;
    }
    
    // Reset retry counter
    (this as any)._chartRetryCount = 0;
    
    this.destroyCharts();
    const isDark = this.themeService.isDarkMode();
    const textColor = isDark ? '#f1f5f9' : '#0f172a';
    const gridColor = isDark ? 'rgba(148,163,184,0.2)' : 'rgba(148,163,184,0.3)';

    // Solves over time (line)
    if (solvesEl && this.analytics.solves_over_time?.length) {
      const c1 = new Chart(solvesEl, {
        type: 'line',
        data: {
          labels: this.analytics.solves_over_time.map(e => e.date.slice(5)),
          datasets: [{ label: 'Solves', data: this.analytics.solves_over_time.map(e => e.count), borderColor: '#22c55e', backgroundColor: 'rgba(34,197,94,0.1)', fill: true, tension: 0.3 }]
        },
        options: { 
          responsive: true, 
          maintainAspectRatio: false, 
          plugins: { 
            legend: { display: false },
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
              ticks: { 
                color: textColor, 
                maxTicksLimit: 10,
                font: { size: 12, weight: 500 }
              },
              grid: { color: gridColor }
            }, 
            y: { 
              beginAtZero: true, 
              ticks: { 
                color: textColor,
                font: { size: 12, weight: 500 }
              }, 
              grid: { color: gridColor } 
            } 
          } 
        }
      });
      this.chartInstances.push(c1);
    }

    // Submissions over time (line)
    const subsEl = document.getElementById('submissionsOverTimeChart') as HTMLCanvasElement;
    if (subsEl && this.analytics.submissions_over_time?.length) {
      const c2 = new Chart(subsEl, {
        type: 'line',
        data: {
          labels: this.analytics.submissions_over_time.map(e => e.date.slice(5)),
          datasets: [{ label: 'Submissions', data: this.analytics.submissions_over_time.map(e => e.count), borderColor: '#3b82f6', backgroundColor: 'rgba(59,130,246,0.1)', fill: true, tension: 0.3 }]
        },
        options: { 
          responsive: true, 
          maintainAspectRatio: false, 
          plugins: { 
            legend: { display: false },
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
              ticks: { 
                color: textColor, 
                maxTicksLimit: 10,
                font: { size: 12, weight: 500 }
              },
              grid: { color: gridColor }
            }, 
            y: { 
              beginAtZero: true, 
              ticks: { 
                color: textColor,
                font: { size: 12, weight: 500 }
              }, 
              grid: { color: gridColor } 
            } 
          } 
        }
      });
      this.chartInstances.push(c2);
    }

    // User growth (line)
    const userGrowthEl = document.getElementById('userGrowthChart') as HTMLCanvasElement;
    if (userGrowthEl && this.analytics.user_growth?.length) {
      const c3 = new Chart(userGrowthEl, {
        type: 'line',
        data: {
          labels: this.analytics.user_growth.map(e => e.date.slice(5)),
          datasets: [{ label: 'New users', data: this.analytics.user_growth.map(e => e.count), borderColor: '#8b5cf6', backgroundColor: 'rgba(139,92,246,0.1)', fill: true, tension: 0.3 }]
        },
        options: { 
          responsive: true, 
          maintainAspectRatio: false, 
          plugins: { 
            legend: { display: false },
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
              ticks: { 
                color: textColor, 
                maxTicksLimit: 10,
                font: { size: 12, weight: 500 }
              },
              grid: { color: gridColor }
            }, 
            y: { 
              beginAtZero: true, 
              ticks: { 
                color: textColor,
                font: { size: 12, weight: 500 }
              }, 
              grid: { color: gridColor } 
            } 
          } 
        }
      });
      this.chartInstances.push(c3);
    }

    // Team growth (line)
    const teamGrowthEl = document.getElementById('teamGrowthChart') as HTMLCanvasElement;
    if (teamGrowthEl && this.analytics.team_growth?.length) {
      const c4 = new Chart(teamGrowthEl, {
        type: 'line',
        data: {
          labels: this.analytics.team_growth.map(e => e.date.slice(5)),
          datasets: [{ label: 'New teams', data: this.analytics.team_growth.map(e => e.count), borderColor: '#f59e0b', backgroundColor: 'rgba(245,158,11,0.1)', fill: true, tension: 0.3 }]
        },
        options: { 
          responsive: true, 
          maintainAspectRatio: false, 
          plugins: { 
            legend: { display: false },
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
              ticks: { 
                color: textColor, 
                maxTicksLimit: 10,
                font: { size: 12, weight: 500 }
              },
              grid: { color: gridColor }
            }, 
            y: { 
              beginAtZero: true, 
              ticks: { 
                color: textColor,
                font: { size: 12, weight: 500 }
              }, 
              grid: { color: gridColor } 
            } 
          } 
        }
      });
      this.chartInstances.push(c4);
    }

    // Category breakdown (doughnut)
    const catEl = document.getElementById('categoryBreakdownChart') as HTMLCanvasElement;
    if (catEl && this.categoryKeys.length) {
      const colors = ['#6366f1', '#22c55e', '#eab308', '#ef4444', '#ec4899', '#14b8a6', '#f97316', '#8b5cf6', '#64748b'];
      const c5 = new Chart(catEl, {
        type: 'doughnut',
        data: {
          labels: this.categoryKeys.map(k => k.charAt(0).toUpperCase() + k.slice(1)),
          datasets: [{ data: this.categoryKeys.map(k => this.analytics!.category_breakdown[k]), backgroundColor: this.categoryKeys.map((_, i) => colors[i % colors.length]) }]
        },
        options: { 
          responsive: true, 
          maintainAspectRatio: false, 
          plugins: { 
            legend: { 
              position: 'right', 
              labels: { 
                color: textColor,
                font: { size: 13, weight: 500 },
                padding: 15
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
          } 
        }
      });
      this.chartInstances.push(c5);
    }

    // Difficulty breakdown (doughnut)
    const diffEl = document.getElementById('difficultyBreakdownChart') as HTMLCanvasElement;
    if (diffEl && this.difficultyKeys.length) {
      const diffColors: Record<string, string> = { easy: '#22c55e', medium: '#eab308', hard: '#ef4444' };
      const c6 = new Chart(diffEl, {
        type: 'doughnut',
        data: {
          labels: this.difficultyKeys.map(k => k.charAt(0).toUpperCase() + k.slice(1)),
          datasets: [{ data: this.difficultyKeys.map(k => this.analytics!.difficulty_breakdown[k]), backgroundColor: this.difficultyKeys.map(k => diffColors[k] || '#64748b') }]
        },
        options: { 
          responsive: true, 
          maintainAspectRatio: false, 
          plugins: { 
            legend: { 
              position: 'right', 
              labels: { 
                color: textColor,
                font: { size: 13, weight: 500 },
                padding: 15
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
          } 
        }
      });
      this.chartInstances.push(c6);
    }

    // Top teams (bar)
    const topTeamsEl = document.getElementById('topTeamsChart') as HTMLCanvasElement;
    if (topTeamsEl && this.analytics.top_teams?.length) {
      const c7 = new Chart(topTeamsEl, {
        type: 'bar',
        data: {
          labels: this.analytics.top_teams.map(t => t.name.length > 12 ? t.name.slice(0, 12) + '…' : t.name),
          datasets: [{ label: 'Score', data: this.analytics.top_teams.map(t => t.score), backgroundColor: 'rgba(34,197,94,0.7)' }]
        },
        options: { 
          indexAxis: 'y', 
          responsive: true, 
          maintainAspectRatio: false, 
          plugins: { 
            legend: { display: false },
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
              beginAtZero: true, 
              ticks: { 
                color: textColor,
                font: { size: 12, weight: 500 }
              }, 
              grid: { color: gridColor } 
            }, 
            y: { 
              ticks: { 
                color: textColor,
                font: { size: 12, weight: 500 }
              },
              grid: { color: gridColor }
            } 
          } 
        }
      });
      this.chartInstances.push(c7);
    }

    // Top users (bar)
    const topUsersEl = document.getElementById('topUsersChart') as HTMLCanvasElement;
    if (topUsersEl && this.analytics.top_users?.length) {
      const c8 = new Chart(topUsersEl, {
        type: 'bar',
        data: {
          labels: this.analytics.top_users.map(u => u.username.length > 12 ? u.username.slice(0, 12) + '…' : u.username),
          datasets: [{ label: 'Score', data: this.analytics.top_users.map(u => u.score), backgroundColor: 'rgba(59,130,246,0.7)' }]
        },
        options: { 
          indexAxis: 'y', 
          responsive: true, 
          maintainAspectRatio: false, 
          plugins: { 
            legend: { display: false },
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
              beginAtZero: true, 
              ticks: { 
                color: textColor,
                font: { size: 12, weight: 500 }
              }, 
              grid: { color: gridColor } 
            }, 
            y: { 
              ticks: { 
                color: textColor,
                font: { size: 12, weight: 500 }
              },
              grid: { color: gridColor }
            } 
          } 
        }
      });
      this.chartInstances.push(c8);
    }

    // Challenge popularity top 10 (bar)
    const popEl = document.getElementById('challengePopularityChart') as HTMLCanvasElement;
    const popData = this.analytics.challenge_popularity?.slice(0, 10) || [];
    if (popEl && popData.length) {
      const c9 = new Chart(popEl, {
        type: 'bar',
        data: {
          labels: popData.map(c => c.title.length > 15 ? c.title.slice(0, 15) + '…' : c.title),
          datasets: [{ label: 'Solves', data: popData.map(c => c.solve_count), backgroundColor: 'rgba(139,92,246,0.7)' }]
        },
        options: { 
          responsive: true, 
          maintainAspectRatio: false, 
          plugins: { 
            legend: { display: false },
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
              ticks: { 
                color: textColor, 
                maxRotation: 45,
                font: { size: 12, weight: 500 }
              },
              grid: { color: gridColor }
            }, 
            y: { 
              beginAtZero: true, 
              ticks: { 
                color: textColor,
                font: { size: 12, weight: 500 }
              }, 
              grid: { color: gridColor } 
            } 
          } 
        }
      });
      this.chartInstances.push(c9);
    }
  }

  ngOnDestroy(): void {
    this.destroyCharts();
  }

  // User Management
  loadUsers(): void {
    this.usersLoading = true;
    this.adminUserService.listUsers().subscribe({
      next: (data) => {
        this.users = data || [];
        this.usersLoading = false;
      },
      error: () => {
        this.usersLoading = false;
        this.showMessage('Failed to load users', 'error');
      }
    });
  }

  updateUserStatus(userId: string, status: string): void {
    let reason = '';
    if (status === 'banned') {
      reason = prompt('Enter ban reason:') || '';
    }
    this.adminUserService.updateUserStatus(userId, status, reason).subscribe({
      next: () => {
        this.showMessage('User status updated', 'success');
        this.loadUsers();
      },
      error: () => this.showMessage('Failed to update user status', 'error')
    });
  }

  updateUserRole(userId: string, role: string): void {
    this.adminUserService.updateUserRole(userId, role).subscribe({
      next: () => {
        this.showMessage('User role updated', 'success');
        this.loadUsers();
      },
      error: () => this.showMessage('Failed to update user role', 'error')
    });
  }

  // Bulk Challenge operations
  exportChallenges(): void {
    this.bulkChallengeService.exportChallenges().subscribe({
      next: (data) => {
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'challenges.json';
        a.click();
        window.URL.revokeObjectURL(url);
        this.showMessage('Challenges exported', 'success');
      },
      error: () => this.showMessage('Failed to export challenges', 'error')
    });
  }

  duplicateChallenge(challengeId: string): void {
    this.bulkChallengeService.duplicateChallenge(challengeId).subscribe({
      next: () => {
        this.showMessage('Challenge duplicated', 'success');
        this.loadChallenges();
      },
      error: () => this.showMessage('Failed to duplicate challenge', 'error')
    });
  }

  renderWriteupContent(writeup: any): string {
    const format = writeup.content_format || 'markdown'; // Default to markdown for backward compatibility
    if (format === 'html') {
      return writeup.content || '';
    } else {
      return this.showdownConverter.makeHtml(writeup.content || '');
    }
  }

  renderChallengeDescription(challenge: ChallengeAdmin): string {
    const format = challenge.description_format || 'markdown'; // Default to markdown for backward compatibility
    if (format === 'html') {
      return challenge.description || '';
    } else {
      return this.showdownConverter.makeHtml(challenge.description || '');
    }
  }

  // Team Management
  loadTeams(): void {
    this.teamsLoading = true;
    this.adminTeamService.listTeams().subscribe({
      next: (data) => {
        this.teams = data || [];
        this.teamsLoading = false;
      },
      error: () => {
        this.teamsLoading = false;
        this.showMessage('Failed to load teams', 'error');
      }
    });
  }

  viewTeamDetails(team: AdminTeam): void {
    this.selectedTeam = this.selectedTeam?.id === team.id ? null : team;
    if (this.selectedTeam) {
      this.teamScoreDelta = 0;
      this.teamScoreReason = '';
    }
  }

  updateTeam(teamId: string, name: string, description: string): void {
    this.adminTeamService.updateTeam(teamId, name, description).subscribe({
      next: () => {
        this.showMessage('Team updated successfully', 'success');
        this.loadTeams();
      },
      error: () => this.showMessage('Failed to update team', 'error')
    });
  }

  applyTeamScoreAdjustment(): void {
    if (!this.selectedTeam) {
      return;
    }
    const delta = Number(this.teamScoreDelta);
    if (!delta || isNaN(delta) || delta === 0) {
      this.showMessage('Please enter a non-zero score delta', 'error');
      return;
    }

    this.adminTeamService.adjustScore(this.selectedTeam.id, delta, this.teamScoreReason || '').subscribe({
      next: () => {
        this.showMessage('Team score adjusted', 'success');
        // Optimistically update selectedTeam and teams list
        this.selectedTeam!.score += delta;
        const idx = this.teams.findIndex(t => t.id === this.selectedTeam!.id);
        if (idx !== -1) {
          this.teams[idx] = { ...this.teams[idx], score: this.teams[idx].score + delta };
        }
        this.teamScoreDelta = 0;
        this.teamScoreReason = '';
      },
      error: () => {
        this.showMessage('Failed to adjust team score', 'error');
      }
    });
  }

  changeTeamLeader(teamId: string, newLeaderId: string): void {
    this.adminTeamService.updateTeamLeader(teamId, newLeaderId).subscribe({
      next: () => {
        this.showMessage('Team leader updated successfully', 'success');
        this.loadTeams();
        this.selectedTeam = null;
      },
      error: () => this.showMessage('Failed to update team leader', 'error')
    });
  }

  removeTeamMember(teamId: string, memberId: string): void {
    this.confirmationModalService.show({
      title: 'Remove Team Member',
      message: 'Are you sure you want to remove this member from the team?',
      confirmText: 'Remove',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        this.adminTeamService.removeMember(teamId, memberId).subscribe({
          next: () => {
            this.showMessage('Member removed from team', 'success');
            this.loadTeams();
            this.selectedTeam = null;
          },
          error: () => this.showMessage('Failed to remove member', 'error')
        });
      }
    });
  }

  deleteTeam(teamId: string, teamName: string): void {
    this.confirmationModalService.show({
      title: 'Delete Team',
      message: `Are you sure you want to delete the team "${teamName}"? This action cannot be undone.`,
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        this.adminTeamService.deleteTeam(teamId).subscribe({
          next: () => {
            this.showMessage('Team deleted successfully', 'success');
            this.loadTeams();
            this.selectedTeam = null;
          },
          error: () => this.showMessage('Failed to delete team', 'error')
        });
      }
    });
  }

  // Enhanced User Management
  viewUserDetails(user: AdminUser): void {
    this.selectedUser = this.selectedUser?.id === user.id ? null : user;
    if (this.selectedUser) {
      this.userScoreDelta = 0;
      this.userScoreReason = '';
    }
  }

  deleteUser(userId: string, username: string): void {
    this.confirmationModalService.show({
      title: 'Delete User',
      message: `Are you sure you want to delete the user "${username}"? This action cannot be undone.`,
      confirmText: 'Delete',
      cancelText: 'Cancel'
    }).pipe(take(1)).subscribe(confirmed => {
      if (confirmed) {
        this.adminUserService.deleteUser(userId).subscribe({
          next: () => {
            this.showMessage('User deleted successfully', 'success');
            this.loadUsers();
            this.selectedUser = null;
          },
          error: () => this.showMessage('Failed to delete user', 'error')
        });
      }
    });
  }

  applyUserScoreAdjustment(): void {
    if (!this.selectedUser) {
      return;
    }
    const delta = Number(this.userScoreDelta);
    if (!delta || isNaN(delta) || delta === 0) {
      this.showMessage('Please enter a non-zero score delta', 'error');
      return;
    }

    this.adminUserService.adjustScore(this.selectedUser.id, delta, this.userScoreReason || '').subscribe({
      next: () => {
        this.showMessage('User score adjusted', 'success');
        this.userScoreDelta = 0;
        this.userScoreReason = '';
      },
      error: () => {
        this.showMessage('Failed to adjust user score', 'error');
      }
    });
  }

  private get apiUrl(): string {
    return environment.apiUrl;
  }
}
