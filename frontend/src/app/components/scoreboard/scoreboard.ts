import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { TeamService, Team } from '../../services/team';
import { ScoreboardService } from '../../services/scoreboard';

interface UserScore {
  username: string;
  score: number;
  team_name?: string;
}

@Component({
  selector: 'app-scoreboard',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './scoreboard.html',
  styleUrls: ['./scoreboard.scss']
})
export class ScoreboardComponent implements OnInit {
  teams: Team[] = [];
  users: UserScore[] = [];
  isLoading = true;
  viewMode: 'teams' | 'individuals' = 'teams';

  constructor(
    private teamService: TeamService,
    private scoreboardService: ScoreboardService
  ) { }

  ngOnInit(): void {
    this.loadTeamScoreboard();
  }

  switchView(mode: 'teams' | 'individuals'): void {
    this.viewMode = mode;
    if (mode === 'teams') {
      this.loadTeamScoreboard();
    } else {
      this.loadIndividualScoreboard();
    }
  }

  loadTeamScoreboard(): void {
    this.isLoading = true;
    this.teamService.getTeamScoreboard().subscribe({
      next: (response) => {
        this.teams = response.teams || [];
        this.isLoading = false;
      },
      error: (err) => {
        console.error(err);
        this.teams = [];
        this.isLoading = false;
      }
    });
  }

  loadIndividualScoreboard(): void {
    this.isLoading = true;
    this.scoreboardService.getScoreboard().subscribe({
      next: (response) => {
        this.users = response || [];
        this.isLoading = false;
      },
      error: (err) => {
        console.error(err);
        this.users = [];
        this.isLoading = false;
      }
    });
  }
}
