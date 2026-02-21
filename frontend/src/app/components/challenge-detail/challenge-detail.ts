import { Component, OnInit, OnDestroy, AfterViewChecked, effect, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { EditorModule, TINYMCE_SCRIPT_SRC } from '@tinymce/tinymce-angular';
import TurndownService from 'turndown';
import Showdown from 'showdown';
import * as Prism from 'prismjs';
import 'prismjs/components/prism-python';
import 'prismjs/components/prism-javascript';
import 'prismjs/components/prism-bash';
import 'prismjs/components/prism-java';
import 'prismjs/components/prism-c';
import 'prismjs/components/prism-cpp';
import 'prismjs/components/prism-sql';
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-markup';
import 'prismjs/components/prism-css';
import { ChallengeService, HintResponse } from '../../services/challenge';
import { ThemeService } from '../../services/theme';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

export interface SolveEntry {
  user_id: string;
  username: string;
  team_id?: string;
  team_name?: string;
  solved_at: string;
}

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
export class ChallengeDetailComponent implements OnInit, OnDestroy, AfterViewChecked {
  private destroyRef = inject(DestroyRef);
  challenge: any;
  renderedDescription = '';
  flagForm: FormGroup;
  writeupForm: FormGroup;
  message = '';
  isCorrect = false;
  isSubmitting = false;
  isSolved = false;
  solvedAt: string | null = null;
  
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
  
  // Solves
  solves: SolveEntry[] = [];
  showSolves = false;
  isLoadingSolves = false;

  // Rate limiting
  isRateLimited = false;
  rateLimitSeconds = 0;
  private rateLimitInterval: any;
  
  // Prism highlighting state
  private needsHighlight = false;
  
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
          background-color: #0f172a;
          color: #e2e8f0;
          padding: 10px;
        }
        a { color: #f87171; text-decoration: underline; }
        code { 
          background-color: #1e293b; 
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
    
    // Show editor again
    setTimeout(() => {
      this.showWriteupEditor = true;
    }, 0);
  }
  
  onWriteupEditorChange(event: any): void {
    this.writeupEditorContent = event.editor.getContent();
  }

  ngAfterViewChecked(): void {
    if (this.needsHighlight) {
      this.needsHighlight = false;
      this.highlightAndWrapCodeBlocks();
    }
  }

  private highlightAndWrapCodeBlocks(): void {
    // Highlight all code blocks with Prism
    setTimeout(() => {
      const codeBlocks = document.querySelectorAll('.challenge-description pre code, .writeup-content pre code');
      codeBlocks.forEach((block) => {
        Prism.highlightElement(block as HTMLElement);
      });

      // Wrap pre blocks with copy button if not already wrapped
      const preBlocks = document.querySelectorAll('.challenge-description pre, .writeup-content pre');
      preBlocks.forEach((pre) => {
        if (pre.parentElement?.classList.contains('code-block-wrapper')) return;
        const wrapper = document.createElement('div');
        wrapper.className = 'code-block-wrapper';
        const btn = document.createElement('button');
        btn.className = 'code-copy-btn';
        btn.textContent = 'Copy';
        btn.addEventListener('click', () => this.copyCode(pre as HTMLElement, btn));
        pre.parentNode?.insertBefore(wrapper, pre);
        wrapper.appendChild(btn);
        wrapper.appendChild(pre);
      });
    }, 0);
  }

  copyCode(preElement: HTMLElement, btn: HTMLButtonElement): void {
    const code = preElement.textContent || '';
    navigator.clipboard.writeText(code).then(() => {
      btn.textContent = 'Copied!';
      btn.classList.add('copied');
      setTimeout(() => {
        btn.textContent = 'Copy';
        btn.classList.remove('copied');
      }, 2000);
    });
  }

  toggleSolves(): void {
    if (this.showSolves) {
      this.showSolves = false;
      return;
    }
    this.loadSolves();
  }

  loadSolves(): void {
    if (!this.challenge) return;
    this.isLoadingSolves = true;
    this.showSolves = true;
    this.challengeService.getChallengeSolves(this.challenge.id).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => {
        this.solves = data || [];
        this.isLoadingSolves = false;
      },
      error: () => {
        this.solves = [];
        this.isLoadingSolves = false;
      }
    });
  }

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.challengeService.getChallenge(id).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
        next: (data) => {
          this.challenge = data;

          // Lock the form if this challenge is already solved
          if (data.is_solved) {
            this.isSolved = true;
            this.isCorrect = true;
            this.flagForm.get('flag')?.disable();
          }

          // Render description based on format
          if (this.challenge.description) {
            const format = this.challenge.description_format || 'markdown';
            
            if (format === 'html') {
              this.renderedDescription = this.challenge.description;
            } else {
              this.renderedDescription = this.markdownConverter.makeHtml(this.challenge.description);
            }
          } else {
            this.renderedDescription = '';
          }
          this.needsHighlight = true;
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
          const format = writeup.content_format || 'markdown';
          return {
            ...writeup,
            renderedContent: format === 'html' 
              ? writeup.content 
              : this.markdownConverter.makeHtml(writeup.content || '')
          };
        });
        this.needsHighlight = true;
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
    if (!this.challenge || this.isRateLimited || this.isSubmitting || this.isSolved) return;
    const flagValue = this.flagForm.getRawValue().flag;
    if (!flagValue) return;

    this.isSubmitting = true;
    this.challengeService.submitFlag(this.challenge.id, flagValue).subscribe({
      next: (res) => {
        this.isSubmitting = false;
        this.message = res.message;
        this.isCorrect = res.correct;

        if (res.correct) {
          this.isSolved = true;
          this.solvedAt = new Date().toISOString();
          this.flagForm.get('flag')?.disable();
          if (res.points) this.challenge.current_points = res.points;
          if (res.solve_count !== undefined) this.challenge.solve_count = res.solve_count;
        }

        if (res.already_solved) {
          this.isSolved = true;
          this.isCorrect = true;
          this.message = 'You (or your team) already solved this challenge!';
          this.flagForm.get('flag')?.disable();
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

  private startRateLimitCooldown(seconds: number): void {
    this.isRateLimited = true;
    this.rateLimitSeconds = seconds;
    this.flagForm.get('flag')?.disable();

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
        // Re-enable input only if not already solved
        if (!this.isSolved) {
          this.flagForm.get('flag')?.enable();
        }
      }
    }, 1000);
  }
}
