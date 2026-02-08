import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ChallengeService, Challenge } from '../../services/challenge';

@Component({
  selector: 'app-challenge-list',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './challenge-list.html',
  styleUrls: ['./challenge-list.scss']
})
export class ChallengeListComponent implements OnInit {
  challenges: Challenge[] = [];
  filteredChallenges: Challenge[] = [];

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

  constructor(private challengeService: ChallengeService) { }

  ngOnInit(): void {
    this.challengeService.getChallenges().subscribe({
      next: (data) => {
        this.challenges = data || [];
        // Extract unique categories
        this.categories = [...new Set(this.challenges.map(c => c.category))].sort();
        // Extract unique tags
        const allTags = this.challenges.flatMap(c => c.tags || []);
        this.tags = [...new Set(allTags)].sort();
        this.applyFilters();
      },
      error: (err) => {
        console.error(err);
        this.challenges = [];
        this.filteredChallenges = [];
      }
    });
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
