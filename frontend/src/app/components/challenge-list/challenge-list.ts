import { Component, DestroyRef, OnInit, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ChallengeService, Challenge } from '../../services/challenge';
import { TeamService } from '../../services/team';
import Showdown from 'showdown';

@Component({
  selector: 'app-challenge-list',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './challenge-list.html',
  styleUrls: ['./challenge-list.scss']
})
export class ChallengeListComponent implements OnInit {
  private teamService = inject(TeamService);
  private destroyRef = inject(DestroyRef);

  challenges: Challenge[] = [];
  filteredChallenges: Challenge[] = [];
  isLoading = true;
  hasTeam = false;

  // Precomputed plain-text previews keyed by challenge id
  plainTextPreviews: Map<string, string> = new Map();

  // Filter state
  searchQuery = '';
  selectedCategory = '';
  selectedDifficulty = '';
  selectedTag = '';
  sortBy = 'title';

  // Available filter options
  categories: string[] = [];
  difficulties = ['easy', 'medium', 'hard'];
  tags: string[] = [];

  // Markdown converter with enhanced configuration
  private markdownConverter = new Showdown.Converter({
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

  constructor(private challengeService: ChallengeService) { }

  ngOnInit(): void {
    this.teamService.currentTeam$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(team => {
      this.hasTeam = team !== null;
    });

    this.isLoading = true;
    this.challengeService.getChallenges().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => {
        this.challenges = data || [];
        // Extract unique categories
        this.categories = [...new Set(this.challenges.map(c => c.category))].sort();
        // Extract unique tags
        const allTags = this.challenges.flatMap(c => c.tags || []);
        this.tags = [...new Set(allTags)].sort();
        // Precompute plain-text previews for all challenges
        this.plainTextPreviews.clear();
        this.challenges.forEach(c => {
          this.plainTextPreviews.set(c.id, this.computePlainTextPreview(c, 150));
        });
        this.applyFilters();
        this.isLoading = false;
      },
      error: (err) => {
        console.error(err);
        this.challenges = [];
        this.filteredChallenges = [];
        this.isLoading = false;
      }
    });
  }

  // Internal helper — called once per challenge at load time
  private computePlainTextPreview(challenge: Challenge, maxLength: number = 150): string {
    if (!challenge.description) return '';

    const format = challenge.description_format || 'markdown'; // Default to markdown for backward compatibility
    let html = '';

    if (format === 'html') {
      // Already HTML
      html = challenge.description;
    } else {
      // Convert markdown to HTML
      html = this.markdownConverter.makeHtml(challenge.description);
    }

    // Strip HTML tags to get plain text
    const tmp = document.createElement('div');
    tmp.innerHTML = html;
    const plainText = tmp.textContent || tmp.innerText || '';
    return plainText.length > maxLength ? plainText.slice(0, maxLength) + '...' : plainText;
  }

  applyFilters(): void {
    let result = [...this.challenges];

    // Search filter
    if (this.searchQuery.trim()) {
      const query = this.searchQuery.toLowerCase();
      result = result.filter(c =>
        c.title.toLowerCase().includes(query) ||
        c.description.toLowerCase().includes(query) ||
        c.category.toLowerCase().includes(query) ||
        (c.tags || []).some(t => t.toLowerCase().includes(query))
      );
    }

    // Category filter
    if (this.selectedCategory) {
      result = result.filter(c => c.category === this.selectedCategory);
    }

    // Difficulty filter
    if (this.selectedDifficulty) {
      result = result.filter(c => c.difficulty === this.selectedDifficulty);
    }

    // Tag filter
    if (this.selectedTag) {
      result = result.filter(c => (c.tags || []).includes(this.selectedTag));
    }

    // Sort
    switch (this.sortBy) {
      case 'points-desc':
        result.sort((a, b) => b.current_points - a.current_points);
        break;
      case 'points-asc':
        result.sort((a, b) => a.current_points - b.current_points);
        break;
      case 'solves':
        result.sort((a, b) => b.solve_count - a.solve_count);
        break;
      case 'title':
      default:
        result.sort((a, b) => a.title.localeCompare(b.title));
        break;
    }

    this.filteredChallenges = result;
  }

  clearFilters(): void {
    this.searchQuery = '';
    this.selectedCategory = '';
    this.selectedDifficulty = '';
    this.selectedTag = '';
    this.sortBy = 'title';
    this.applyFilters();
  }
}
