import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { EditorModule, TINYMCE_SCRIPT_SRC } from '@tinymce/tinymce-angular';
import TurndownService from 'turndown';
import Showdown from 'showdown';
import { ChallengeService, ChallengeAdmin, ChallengeRequest } from '../../services/challenge';
import { NotificationService, Notification } from '../../services/notification';

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
  message = '';
  messageType: 'success' | 'error' = 'success';
  
  // Tab state
  activeTab: 'create' | 'manage' | 'notifications' = 'create';
  
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
  
  // Rich text editor content (HTML)
  editorContent = '';
  
  // TinyMCE configuration
  editorConfig: any = {
    base_url: '/tinymce',
    suffix: '.min',
    height: 400,
    menubar: false,
    plugins: [
      'advlist', 'autolink', 'lists', 'link', 'image', 'charmap',
      'anchor', 'searchreplace', 'visualblocks', 'code', 'fullscreen',
      'insertdatetime', 'media', 'table', 'preview', 'help', 'wordcount'
    ],
    toolbar: 'undo redo | blocks | bold italic forecolor backcolor | alignleft aligncenter alignright alignjustify | bullist numlist outdent indent | removeformat | code | help',
    content_style: `
      body { 
        font-family: 'Space Grotesk', Arial, sans-serif; 
        font-size: 14px; 
        background-color: #1e293b;
        color: #e2e8f0;
        padding: 10px;
      }
      a { color: #f87171; }
      code { background-color: #0f172a; padding: 2px 6px; border-radius: 4px; }
      pre { background-color: #0f172a; padding: 12px; border-radius: 8px; overflow-x: auto; }
    `,
    skin: 'oxide-dark',
    content_css: 'dark'
  };
  
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

  // Markdown converter
  private turndownService = new TurndownService();
  private showdownConverter = new Showdown.Converter();

  constructor(
    private fb: FormBuilder,
    private challengeService: ChallengeService,
    private notificationService: NotificationService
  ) {
    this.challengeForm = this.fb.group({
      title: ['', Validators.required],
      category: ['', Validators.required],
      difficulty: ['', Validators.required],
      max_points: [500, [Validators.required, Validators.min(1)]],
      min_points: [100, [Validators.required, Validators.min(1)]],
      decay: [10, [Validators.required, Validators.min(1)]],
      flag: ['', Validators.required],
      files: ['']
    });

    this.notificationForm = this.fb.group({
      title: ['', Validators.required],
      content: ['', Validators.required],
      type: ['info', Validators.required]
    });
  }

  ngOnInit(): void {
    this.loadChallenges();
  }

  loadChallenges(): void {
    this.isLoading = true;
    this.challengeService.getChallengesForAdmin().subscribe({
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

  switchTab(tab: 'create' | 'manage' | 'notifications'): void {
    this.activeTab = tab;
    if (tab === 'manage') {
      this.loadChallenges();
    }
    if (tab === 'notifications') {
      this.loadNotifications();
    }
    if (tab === 'create' && !this.isEditMode) {
      this.resetForm();
    }
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
      // Convert HTML to Markdown for storage
      const markdownDescription = this.turndownService.turndown(this.editorContent);
      
      const formValue = this.challengeForm.value;
      const challenge: ChallengeRequest = {
        title: formValue.title,
        description: markdownDescription,
        category: formValue.category,
        difficulty: formValue.difficulty,
        max_points: formValue.max_points,
        min_points: formValue.min_points,
        decay: formValue.decay,
        flag: formValue.flag,
        files: formValue.files ? formValue.files.split(',').map((f: string) => f.trim()).filter((f: string) => f) : []
      };

      if (this.isEditMode && this.editingChallengeId) {
        // Update existing challenge
        this.challengeService.updateChallenge(this.editingChallengeId, challenge).subscribe({
          next: () => {
            this.showMessage('Challenge updated successfully', 'success');
            this.loadChallenges();
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

  editChallenge(challenge: ChallengeAdmin): void {
    this.isEditMode = true;
    this.editingChallengeId = challenge.id;
    
    // Convert Markdown to HTML for the editor
    this.editorContent = this.showdownConverter.makeHtml(challenge.description);
    
    this.challengeForm.patchValue({
      title: challenge.title,
      category: challenge.category,
      difficulty: challenge.difficulty,
      max_points: challenge.max_points,
      min_points: challenge.min_points,
      decay: challenge.decay,
      flag: '', // Flag is not returned from API for security
      files: challenge.files ? challenge.files.join(', ') : ''
    });
    
    this.switchTab('create');
    this.showMessage(`Editing: ${challenge.title} (Enter the flag again)`, 'success');
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
    this.challengeForm.reset({
      title: '',
      category: '',
      difficulty: '',
      max_points: 500,
      min_points: 100,
      decay: 10,
      flag: '',
      files: ''
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
}
