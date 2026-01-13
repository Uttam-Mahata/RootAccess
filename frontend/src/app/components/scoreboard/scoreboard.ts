import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule } from '@angular/material/table';
import { ScoreboardService } from '../../services/scoreboard';

@Component({
  selector: 'app-scoreboard',
  standalone: true,
  imports: [CommonModule, MatTableModule],
  templateUrl: './scoreboard.html',
  styleUrls: ['./scoreboard.scss']
})
export class ScoreboardComponent implements OnInit {
  scores: any[] = [];
  displayedColumns: string[] = ['rank', 'username', 'score'];

  constructor(private scoreboardService: ScoreboardService) { }

  ngOnInit(): void {
    this.scoreboardService.getScoreboard().subscribe({
      next: (data) => {
        this.scores = data.sort((a, b) => b.score - a.score);
      },
      error: (err) => console.error(err)
    });
  }
}
