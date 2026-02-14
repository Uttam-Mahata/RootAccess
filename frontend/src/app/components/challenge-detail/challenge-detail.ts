import { Component, OnInit, OnDestroy, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { EditorModule, TINYMCE_SCRIPT_SRC } from '@tinymce/tinymce-angular';
import TurndownService from 'turndown';
import Showdown from 'showdown';
import { ChallengeService, HintResponse } from '../../services/challenge';
import { ThemeService } from '../../services/theme';

@Component({
  selector: 'app-challenge-detail',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    ReactiveFormsModule,
    EditorModule
  ],
  providers: [
    { provide: TINYMCE_SCRIPT_SRC, useValue: 'tinymce/tinymce.min.js' }
  ],
  templateUrl: './challenge-detail.html',
  styleUrls: ['./challenge-detail.scss']
})
export class ChallengeDetailComponent implements OnInit, OnDestroy {
  challenge: any;
  renderedDescription = '';
  flagForm: FormGroup;
  writeupForm: FormGroup;
  message = '';
  isCorrect = false;
  isSubmitting = false;
  
  // Hints
  hints: HintResponse[] = [];
  revealingHint: string | null = null;

  // Writeups
  writeups: any[] = [];
  showWriteupForm = false;
  writeupMessage = '';
  isSubmittingWriteup = false;
  writeupEditorContent = '';
  writeupEditorConfig: any = {};
  showWriteupEditor = true;
  
  // Rate limiting
  isRateLimited = false;
  rateLimitSeconds = 0;
  private rateLimitInterval: any;
  
  // Markdown converter with enhanced configuration
  private markdownConverter = new Showdown.Converter({
    tables: true,
    strikethrough: true,
    tasklists: true,
    smoothLivePreview: true,
    simpleLineBreaks: false,
    openLinksInNewWindow: true,
    emoji: true,
    ghCodeBlocks: true,
    ghCompatibleHeaderId: true,
    encodeEmails: true,
    simplifiedAutoLink: true,
    literalMidWordUnderscores: true,
    parseImgDimensions: true,
    requireSpaceBeforeHeadingText: false,
    // Enable GitHub Flavored Markdown extensions
    extensions: []
  });
  
  // Markdown to HTML converter for writeup submission
  private turndownService = new TurndownService();

  constructor(
    private route: ActivatedRoute,
    private challengeService: ChallengeService,
    private fb: FormBuilder,
    private themeService: ThemeService
  ) {
    this.flagForm = this.fb.group({
      flag: ['', Validators.required]
    });
    this.writeupForm = this.fb.group({
      content: ['', [Validators.required, Validators.minLength(50)]],
      content_format: ['markdown', Validators.required]
    });
    
    // Initialize writeup editor config
    this.updateWriteupEditorConfig();
    
    // Watch for theme changes and update editor config
    effect(() => {
      this.themeService.isDarkMode();
      this.updateWriteupEditorConfig();
    });
  }
  
  private updateWriteupEditorConfig(): void {
    const isDark = this.themeService.isDarkMode();
    
    // Temporarily hide editor to force re-render with new theme
    this.showWriteupEditor = false;
    
    this.writeupEditorConfig = {
      base_url: '/tinymce',
      suffix: '.min',
      height: 350,
      menubar: false,
      branding: false,
      promotion: false,
      plugins: [
        'advlist', 'autolink', 'lists', 'link', 'charmap',
        'searchreplace', 'visualblocks', 'code', 'codesample',
        'insertdatetime', 'table', 'help', 'wordcount'
      ],
      toolbar: 'undo redo | blocks | bold italic | bullist numlist | codesample code | removeformat | help',
      codesample_languages: [
        { text: 'HTML/XML', value: 'markup' },
        { text: 'JavaScript', value: 'javascript' },
        { text: 'Python', value: 'python' },
        { text: 'Java', value: 'java' },
        { text: 'C', value: 'c' },
        { text: 'C++', value: 'cpp' },
        { text: 'Bash', value: 'bash' },
        { text: 'SQL', value: 'sql' }
      ],
      content_style: isDark ? `
        body { 
          font-family: 'Space Grotesk', Arial, sans-serif; 
          font-size: 14px; 
          background-color: #1e293b;
          color: #e2e8f0;
          padding: 10px;
        }
        a { color: #f87171; }
        code { background-color: #0f172a; padding: 3px 8px; border-radius: 4px; color: #fbbf24; }
        pre { background-color: #0f172a; padding: 16px; border-radius: 8px; color: #e2e8f0; }
      ` : `
        body { 
          font-family: 'Space Grotesk', Arial, sans-serif; 
          font-size: 14px; 
          background-color: #ffffff;
          color: #1e293b;
          padding: 10px;
        }
        a { color: #dc2626; }
        code { background-color: #f1f5f9; padding: 3px 8px; border-radius: 4px; color: #b91c1c; }
        pre { background-color: #f1f5f9; padding: 16px; border-radius: 8px; color: #1e293b; }
      `,
      skin: isDark ? 'oxide-dark' : 'oxide',
      content_css: isDark ? 'dark' : 'default'
    };
    
    // Show editor again
    setTimeout(() => {
      this.showWriteupEditor = true;
    }, 0);
  }
  
  onWriteupEditorChange(event: any): void {
    this.writeupEditorContent = event.editor.getContent();
  }

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.challengeService.getChallenge(id).subscribe({
        next: (data) => {
          this.challenge = data;
          console.log('Challenge data:', this.challenge);
          console.log('Description format:', this.challenge.description_format);
          console.log('Description (first 200 chars):', this.challenge.description?.substring(0, 200));
          
          // Render description based on format
          if (this.challenge.description) {
            const format = this.challenge.description_format || 'markdown'; // Default to markdown for backward compatibility
            console.log('Using format:', format);
            
            if (format === 'html') {
              // Already HTML, use directly
              this.renderedDescription = this.challenge.description;
            } else {
              // Convert markdown to HTML
              this.renderedDescription = this.markdownConverter.makeHtml(this.challenge.description);
            }
            
            console.log('Rendered description (first 200 chars):', this.renderedDescription.substring(0, 200));
          } else {
            this.renderedDescription = '';
          }
          this.loadHints();
          this.loadWriteups();
        },
        error: (err) => console.error('Error loading challenge:', err)
      });
    }
  }

  ngOnDestroy(): void {
    if (this.rateLimitInterval) {
      clearInterval(this.rateLimitInterval);
    }
  }

  loadHints(): void {
    if (!this.challenge) return;
    this.challengeService.getHints(this.challenge.id).subscribe({
      next: (hints) => this.hints = hints || [],
      error: () => this.hints = []
    });
  }

  revealHint(hintId: string): void {
    if (!this.challenge || this.revealingHint) return;
    
    const hint = this.hints.find(h => h.id === hintId);
    if (!hint || hint.revealed) return;

    if (!confirm(`Revealing this hint will cost ${hint.cost} points. Are you sure?`)) return;

    this.revealingHint = hintId;
    this.challengeService.revealHint(this.challenge.id, hintId).subscribe({
      next: (revealed) => {
        const index = this.hints.findIndex(h => h.id === hintId);
        if (index !== -1) {
          this.hints[index] = revealed;
        }
        this.revealingHint = null;
      },
      error: (err) => {
        console.error('Error revealing hint:', err);
        this.revealingHint = null;
      }
    });
  }

  loadWriteups(): void {
    if (!this.challenge) return;
    this.challengeService.getWriteups(this.challenge.id).subscribe({
      next: (writeups) => {
        this.writeups = (writeups || []).map(writeup => {
          const format = writeup.content_format || 'markdown'; // Default to markdown for backward compatibility
          return {
            ...writeup,
            renderedContent: format === 'html' 
              ? writeup.content 
              : this.markdownConverter.makeHtml(writeup.content || '')
          };
        });
      },
      error: () => this.writeups = []
    });
  }

  submitWriteup(): void {
    if (!this.challenge || this.isSubmittingWriteup || !this.writeupEditorContent.trim()) return;

    const formValue = this.writeupForm.value;
    const selectedFormat = formValue.content_format || 'markdown';
    
    // Convert based on selected format
    let content: string;
    if (selectedFormat === 'markdown') {
      // Convert TinyMCE HTML to Markdown
      content = this.turndownService.turndown(this.writeupEditorContent);
    } else {
      // Store as HTML directly
      content = this.writeupEditorContent;
    }

    this.isSubmittingWriteup = true;
    this.challengeService.submitWriteup(this.challenge.id, content, selectedFormat).subscribe({
      next: (res) => {
        this.writeupMessage = res.message || 'Writeup submitted for review!';
        this.showWriteupForm = false;
        this.writeupForm.reset({ content_format: 'markdown' });
        this.writeupEditorContent = '';
        this.isSubmittingWriteup = false;
        this.loadWriteups();
      },
      error: (err) => {
        this.writeupMessage = err.error?.error || 'Error submitting writeup';
        this.isSubmittingWriteup = false;
      }
    });
  }

  onSubmit(): void {
    if (this.flagForm.valid && this.challenge && !this.isRateLimited && !this.isSubmitting) {
      this.isSubmitting = true;
      this.challengeService.submitFlag(this.challenge.id, this.flagForm.value.flag).subscribe({
        next: (res) => {
          this.message = res.message;
          this.isCorrect = res.correct;
          this.isSubmitting = false;
          
          // Update challenge points if correct
          if (res.correct && res.points) {
            this.challenge.current_points = res.points;
          }
          if (res.solve_count !== undefined) {
            this.challenge.solve_count = res.solve_count;
          }
        },
        error: (err) => {
          this.isSubmitting = false;
          
          // Handle rate limiting (429 Too Many Requests)
          if (err.status === 429) {
            const retryAfter = err.error?.retry_after || 60;
            this.startRateLimitCooldown(retryAfter);
            this.message = `Too many attempts! Please wait ${retryAfter} seconds before trying again.`;
            this.isCorrect = false;
          } else {
            this.message = err.error?.error || 'Error submitting flag';
            this.isCorrect = false;
          }
        }
      });
    }
  }

  private startRateLimitCooldown(seconds: number): void {
    this.isRateLimited = true;
    this.rateLimitSeconds = seconds;

    // Clear any existing interval
    if (this.rateLimitInterval) {
      clearInterval(this.rateLimitInterval);
    }

    this.rateLimitInterval = setInterval(() => {
      this.rateLimitSeconds--;
      if (this.rateLimitSeconds <= 0) {
        this.isRateLimited = false;
        this.rateLimitSeconds = 0;
        this.message = '';
        clearInterval(this.rateLimitInterval);
      }
    }, 1000);
  }
}
