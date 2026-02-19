import { Component, OnInit, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, FormArray } from '@angular/forms';
import { EditorModule, TINYMCE_SCRIPT_SRC } from '@tinymce/tinymce-angular';
import TurndownService from 'turndown';
import Showdown from 'showdown';
import { ChallengeService, ChallengeAdmin, ChallengeRequest, HintRequest } from '../../services/challenge';
import { NotificationService, Notification } from '../../services/notification';
import { ContestService } from '../../services/contest';
import { AnalyticsService, AdminAnalytics } from '../../services/analytics';
import { AdminUserService, AdminUser } from '../../services/admin-user';
import { AdminTeamService, AdminTeam } from '../../services/admin-team';
import { BulkChallengeService } from '../../services/bulk-challenge';
import { ThemeService } from '../../services/theme';
import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-admin-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    EditorModule
  ],
  providers: [
    { provide: TINYMCE_SCRIPT_SRC, useValue: 'tinymce/tinymce.min.js' }
  ],
  templateUrl: './admin-dashboard.html',
  styleUrls: ['./admin-dashboard.scss']
})
export class AdminDashboardComponent implements OnInit {
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

  // Writeups
  writeups: any[] = [];
  isLoadingWriteups = false;

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

  // User Management
  users: AdminUser[] = [];
  usersLoading = false;
  selectedUser: AdminUser | null = null;

  // Team Management
  teams: AdminTeam[] = [];
  teamsLoading = false;
  selectedTeam: AdminTeam | null = null;

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
    private analyticsService: AnalyticsService,
    private adminUserService: AdminUserService,
    private adminTeamService: AdminTeamService,
    private bulkChallengeService: BulkChallengeService,
    private http: HttpClient,
    private themeService: ThemeService
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
    this.loadChallenges();
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
    this.activeTab = tab;
    this.mobileMenuOpen = false; // Close mobile menu on tab switch
    if (tab === 'manage') {
      this.loadChallenges();
    }
    if (tab === 'notifications') {
      this.loadNotifications();
    }
    if (tab === 'contest') {
      this.loadContestConfig();
    }
    if (tab === 'writeups') {
      this.loadWriteups();
    }
    if (tab === 'audit') {
      this.loadAuditLogs();
    }
    if (tab === 'analytics') {
      this.loadAnalytics();
    }
    if (tab === 'users') {
      this.loadUsers();
    }
    if (tab === 'teams') {
      this.loadTeams();
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
    if (confirm(`Are you sure you want to delete the notification "${notification.title}"?`)) {
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
    if (this.previewChallenge?.id === challenge.id) {
      this.previewChallenge = null;
      return;
    }
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
    if (confirm(`Are you sure you want to delete "${challenge.title}"? This action cannot be undone.`)) {
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
        this.isLoadingContest = false;
      }
    });
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

  getScoringTypeLabel(value: string): string {
    const type = this.scoringTypes.find(t => t.value === value);
    return type ? type.label : value || 'Dynamic';
  }

  // Writeup management
  loadWriteups(): void {
    this.isLoadingWriteups = true;
    this.http.get<any[]>(`${this.apiUrl}/admin/writeups`).subscribe({
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

  approveWriteup(id: string): void {
    this.http.put<any>(`${this.apiUrl}/admin/writeups/${id}/status`, { status: 'approved' }).subscribe({
      next: () => {
        this.showMessage('Writeup approved', 'success');
        this.loadWriteups();
      },
      error: () => this.showMessage('Error approving writeup', 'error')
    });
  }

  rejectWriteup(id: string): void {
    this.http.put<any>(`${this.apiUrl}/admin/writeups/${id}/status`, { status: 'rejected' }).subscribe({
      next: () => {
        this.showMessage('Writeup rejected', 'success');
        this.loadWriteups();
      },
      error: () => this.showMessage('Error rejecting writeup', 'error')
    });
  }

  deleteWriteup(id: string): void {
    if (confirm('Are you sure you want to delete this writeup?')) {
      this.http.delete<any>(`${this.apiUrl}/admin/writeups/${id}`).subscribe({
        next: () => {
          this.showMessage('Writeup deleted', 'success');
          this.loadWriteups();
        },
        error: () => this.showMessage('Error deleting writeup', 'error')
      });
    }
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
    this.analyticsService.getPlatformAnalytics().subscribe({
      next: (data) => {
        this.analytics = data;
        this.categoryKeys = data.category_breakdown ? Object.keys(data.category_breakdown) : [];
        this.difficultyKeys = data.difficulty_breakdown ? Object.keys(data.difficulty_breakdown) : [];
        this.analyticsLoading = false;
      },
      error: () => {
        this.analyticsLoading = false;
        this.showMessage('Failed to load analytics', 'error');
      }
    });
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
    if (confirm('Are you sure you want to remove this member from the team?')) {
      this.adminTeamService.removeMember(teamId, memberId).subscribe({
        next: () => {
          this.showMessage('Member removed from team', 'success');
          this.loadTeams();
          this.selectedTeam = null;
        },
        error: () => this.showMessage('Failed to remove member', 'error')
      });
    }
  }

  deleteTeam(teamId: string, teamName: string): void {
    if (confirm(`Are you sure you want to delete the team "${teamName}"? This action cannot be undone.`)) {
      this.adminTeamService.deleteTeam(teamId).subscribe({
        next: () => {
          this.showMessage('Team deleted successfully', 'success');
          this.loadTeams();
          this.selectedTeam = null;
        },
        error: () => this.showMessage('Failed to delete team', 'error')
      });
    }
  }

  // Enhanced User Management
  viewUserDetails(user: AdminUser): void {
    this.selectedUser = this.selectedUser?.id === user.id ? null : user;
  }

  deleteUser(userId: string, username: string): void {
    if (confirm(`Are you sure you want to delete the user "${username}"? This action cannot be undone.`)) {
      this.adminUserService.deleteUser(userId).subscribe({
        next: () => {
          this.showMessage('User deleted successfully', 'success');
          this.loadUsers();
          this.selectedUser = null;
        },
        error: () => this.showMessage('Failed to delete user', 'error')
      });
    }
  }

  private get apiUrl(): string {
    return environment.apiUrl;
  }
}
