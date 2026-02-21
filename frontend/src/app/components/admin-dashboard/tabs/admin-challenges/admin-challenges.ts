import { Component, OnInit, Input, Output, EventEmitter, effect, ChangeDetectorRef, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { EditorModule, TINYMCE_SCRIPT_SRC } from '@tinymce/tinymce-angular';
import TurndownService from 'turndown';
import Showdown from 'showdown';
import { ChallengeService, ChallengeAdmin, ChallengeRequest } from '../../../../services/challenge';
import { BulkChallengeService } from '../../../../services/bulk-challenge';
import { ThemeService } from '../../../../services/theme';
import { ConfirmationModalService } from '../../../../services/confirmation-modal.service';
import { take } from 'rxjs/operators';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

@Component({
  selector: 'app-admin-challenges',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, EditorModule],
  providers: [
    { provide: TINYMCE_SCRIPT_SRC, useValue: 'tinymce/tinymce.min.js' }
  ],
  templateUrl: './admin-challenges.html',
  styleUrls: ['./admin-challenges.scss']
})
export class AdminChallengesComponent implements OnInit {
  private destroyRef = inject(DestroyRef);
  private challengeService = inject(ChallengeService);
  private bulkChallengeService = inject(BulkChallengeService);
  private themeService = inject(ThemeService);
  private confirmationModalService = inject(ConfirmationModalService);
  private cdr = inject(ChangeDetectorRef);
  private fb = inject(FormBuilder);

  @Input() initialView: 'create' | 'manage' = 'manage';
  @Output() viewChanged = new EventEmitter<'create' | 'manage'>();
  @Output() isEditModeChanged = new EventEmitter<boolean>();
  @Output() countChanged = new EventEmitter<number>();
  @Output() messageEmitted = new EventEmitter<{ msg: string; type: 'success' | 'error' }>();

  activeView: 'create' | 'manage' = 'manage';

  challengeForm: FormGroup;
  challenges: ChallengeAdmin[] = [];
  isLoading = false;
  isEditMode = false;
  editingChallengeId: string | null = null;
  previewChallenge: ChallengeAdmin | null = null;

  editorContent = '';
  editorConfig: any = {};
  editorKey = 0;
  showEditor = true;

  scoringTypes = [
    { value: 'dynamic', label: 'Dynamic (CTFd Formula)' },
    { value: 'linear', label: 'Linear Decay' },
    { value: 'static', label: 'Static (Fixed Points)' }
  ];

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

  difficulties = [
    { value: 'easy', label: 'Easy', color: 'green' },
    { value: 'medium', label: 'Medium', color: 'yellow' },
    { value: 'hard', label: 'Hard', color: 'red' }
  ];

  private turndownService = new TurndownService();
  private showdownConverter = new Showdown.Converter({
    tables: true,
    strikethrough: true,
    tasklists: true,
    smoothLivePreview: true,
    simpleLineBreaks: false,
    openLinksInNewWindow: true,
    emoji: true,
    ghCodeBlocks: true,
    encodeEmails: true,
    simplifiedAutoLink: true,
    literalMidWordUnderscores: true,
    parseImgDimensions: true
  });

  constructor() {
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

    this.updateEditorConfig();

    effect(() => {
      this.themeService.isDarkMode();
      this.updateEditorConfig();
    });
  }

  ngOnInit(): void {
    this.activeView = this.initialView;
    this.loadChallenges();
  }

  private updateEditorConfig(): void {
    const isDark = this.themeService.isDarkMode();
    this.showEditor = false;
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
    setTimeout(() => { this.showEditor = true; }, 0);
  }

  switchView(view: 'create' | 'manage'): void {
    if (this.activeView === view) return;
    this.activeView = view;
    this.cdr.markForCheck();
    this.viewChanged.emit(view);
    if (view === 'create' && !this.isEditMode) {
      this.resetForm();
    }
  }

  loadChallenges(): void {
    this.isLoading = true;
    this.challengeService.getChallengesForAdmin(true).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => {
        this.challenges = data || [];
        this.isLoading = false;
        this.countChanged.emit(this.challenges.length);
      },
      error: () => {
        this.challenges = [];
        this.isLoading = false;
        this.messageEmitted.emit({ msg: 'Error loading challenges', type: 'error' });
      }
    });
  }

  onEditorChange(event: any): void {
    this.editorContent = event.editor.getContent();
  }

  onSubmit(): void {
    if (this.challengeForm.valid && this.editorContent.trim()) {
      const formValue = this.challengeForm.value;
      const selectedFormat = formValue.description_format || 'markdown';

      let description: string;
      if (selectedFormat === 'markdown') {
        description = this.turndownService.turndown(this.editorContent);
      } else {
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
        this.challengeService.updateChallenge(this.editingChallengeId, challenge).subscribe({
          next: () => {
            this.messageEmitted.emit({ msg: 'Challenge updated successfully', type: 'success' });
            this.loadChallenges();
            this.resetForm();
            this.switchView('manage');
          },
          error: () => {
            this.messageEmitted.emit({ msg: 'Error updating challenge', type: 'error' });
          }
        });
      } else {
        this.challengeService.createChallenge(challenge).subscribe({
          next: () => {
            this.messageEmitted.emit({ msg: 'Challenge created successfully', type: 'success' });
            this.loadChallenges();
            this.resetForm();
          },
          error: () => {
            this.messageEmitted.emit({ msg: 'Error creating challenge', type: 'error' });
          }
        });
      }
    } else if (!this.editorContent.trim()) {
      this.messageEmitted.emit({ msg: 'Please provide a description for the challenge', type: 'error' });
    }
  }

  previewChallengeToggle(challenge: ChallengeAdmin): void {
    if (!challenge.description) {
      this.challengeService.getChallenge(challenge.id).subscribe({
        next: (full) => {
          this.previewChallenge = { ...challenge, description: full.description, description_format: full.description_format };
        },
        error: () => this.messageEmitted.emit({ msg: 'Failed to load challenge details', type: 'error' })
      });
    } else {
      this.previewChallenge = challenge;
    }
  }

  editChallenge(challenge: ChallengeAdmin): void {
    this.isEditMode = true;
    this.editingChallengeId = challenge.id;
    this.isEditModeChanged.emit(true);

    const applyEdit = (ch: ChallengeAdmin) => {
      const format = ch.description_format || 'markdown';
      if (format === 'html') {
        this.editorContent = ch.description;
      } else {
        this.editorContent = this.showdownConverter.makeHtml(ch.description || '');
      }
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
      this.switchView('create');
      this.messageEmitted.emit({ msg: `Editing: ${ch.title} (Leave flag empty to keep current flag)`, type: 'success' });
    };

    if (!challenge.description) {
      this.challengeService.getChallenge(challenge.id).subscribe({
        next: (full) => {
          applyEdit({ ...challenge, description: full.description, description_format: full.description_format });
        },
        error: () => this.messageEmitted.emit({ msg: 'Failed to load challenge for edit', type: 'error' })
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
            this.messageEmitted.emit({ msg: 'Challenge deleted successfully', type: 'success' });
            this.loadChallenges();
          },
          error: () => {
            this.messageEmitted.emit({ msg: 'Error deleting challenge', type: 'error' });
          }
        });
      }
    });
  }

  resetForm(): void {
    this.isEditMode = false;
    this.editingChallengeId = null;
    this.editorContent = '';
    this.isEditModeChanged.emit(false);
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
  }

  cancelEdit(): void {
    this.resetForm();
    this.switchView('manage');
  }

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
        this.messageEmitted.emit({ msg: 'Challenges exported', type: 'success' });
      },
      error: () => this.messageEmitted.emit({ msg: 'Failed to export challenges', type: 'error' })
    });
  }

  duplicateChallenge(challengeId: string): void {
    this.bulkChallengeService.duplicateChallenge(challengeId).subscribe({
      next: () => {
        this.messageEmitted.emit({ msg: 'Challenge duplicated', type: 'success' });
        this.loadChallenges();
      },
      error: () => this.messageEmitted.emit({ msg: 'Failed to duplicate challenge', type: 'error' })
    });
  }

  getCategoryLabel(value: string): string {
    const category = this.categories.find(c => c.value === value);
    return category ? category.label : value;
  }

  getDifficultyLabel(value: string): string {
    const difficulty = this.difficulties.find(d => d.value === value);
    return difficulty ? difficulty.label : value;
  }

  renderChallengeDescription(challenge: ChallengeAdmin): string {
    const format = challenge.description_format || 'markdown';
    if (format === 'html') {
      return challenge.description || '';
    } else {
      return this.showdownConverter.makeHtml(challenge.description || '');
    }
  }
}
