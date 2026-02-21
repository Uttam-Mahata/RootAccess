import { Component, OnInit, Output, EventEmitter, inject, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { take } from 'rxjs/operators';
import { ContestService } from '../../../../services/contest';
import { ContestAdminService, Contest, ContestRound } from '../../../../services/contest-admin';
import { ChallengeService, ChallengeAdmin } from '../../../../services/challenge';
import { ConfirmationModalService } from '../../../../services/confirmation-modal.service';
import { DatetimePickerComponent } from '../../../datetime-picker/datetime-picker';

@Component({
  selector: 'app-admin-contest',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, DatetimePickerComponent],
  templateUrl: './admin-contest.html',
  styleUrls: ['./admin-contest.scss']
})
export class AdminContestComponent implements OnInit {
  private destroyRef = inject(DestroyRef);
  private contestService = inject(ContestService);
  private contestAdminService = inject(ContestAdminService);
  private challengeService = inject(ChallengeService);
  private fb = inject(FormBuilder);
  private confirmationModalService = inject(ConfirmationModalService);

  @Output() messageEmitted = new EventEmitter<{ msg: string; type: 'success' | 'error' }>();

  // Contest state
  contestConfig: any = null;
  isLoadingContest = false;
  isLoadingContests = false;
  contests: Contest[] = [];
  selectedContest: Contest | null = null;
  activeContestId: string | null = null;
  isCreatingContest = false;
  isEditingContest = false;
  editingContestId: string | null = null;
  contestEntityForm: FormGroup;

  // Round state
  isLoadingRounds = false;
  rounds: ContestRound[] = [];
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
  challenges: ChallengeAdmin[] = [];

  // Categories for labels
  private categories = [
    { value: 'web', label: 'Web Exploitation' }, { value: 'crypto', label: 'Cryptography' },
    { value: 'pwn', label: 'Binary Exploitation (Pwn)' }, { value: 'reverse', label: 'Reverse Engineering' },
    { value: 'forensics', label: 'Digital Forensics' }, { value: 'networking', label: 'Networking' },
    { value: 'steganography', label: 'Steganography' }, { value: 'osint', label: 'OSINT' },
    { value: 'misc', label: 'General Skills/Misc' }
  ];

  constructor() {
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
  }

  ngOnInit(): void {
    this.loadContestConfig();
    this.loadContests();
    this.loadChallenges();
  }

  loadChallenges(): void {
    this.challengeService.getChallengesForAdmin(true).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => {
        this.challenges = data || [];
        if (this.selectedRound) {
          this.computeAvailableChallenges();
        }
      },
      error: () => {}
    });
  }

  loadContestConfig(): void {
    this.isLoadingContest = true;
    this.contestService.getContestConfig().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => {
        this.contestConfig = data.config || null;
        this.activeContestId = this.contestConfig?.contest_id || null;
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
    this.contestAdminService.listContests().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => { this.contests = data || []; this.isLoadingContests = false; },
      error: () => { this.contests = []; this.isLoadingContests = false; }
    });
  }

  loadRounds(): void {
    if (!this.selectedContest?.id) return;
    this.isLoadingRounds = true;
    this.contestAdminService.listRounds(this.selectedContest.id).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (data) => { this.rounds = (data || []).sort((a, b) => a.order - b.order); this.isLoadingRounds = false; },
      error: () => { this.rounds = []; this.isLoadingRounds = false; }
    });
  }

  selectContest(contest: Contest): void {
    this.selectedContest = contest;
    this.loadRounds();
    this.isEditingContest = false; this.editingContestId = null;
    this.isCreatingRound = false; this.isEditingRound = false; this.editingRoundId = null;
    this.selectedRound = null; this.roundChallengeIds = []; this.availableChallengesForRound = [];
  }

  startCreateContest(): void {
    this.isCreatingContest = true; this.isEditingContest = false; this.editingContestId = null;
    this.contestEntityForm.reset({ name: '', description: '', start_time: '', end_time: '' });
  }

  startEditContest(contest: Contest): void {
    this.isCreatingContest = false; this.isEditingContest = true; this.editingContestId = contest.id;
    this.contestEntityForm.patchValue({
      name: contest.name, description: contest.description || '',
      start_time: contest.start_time ? new Date(contest.start_time).toISOString() : '',
      end_time: contest.end_time ? new Date(contest.end_time).toISOString() : ''
    });
  }

  cancelContestForm(): void {
    this.isCreatingContest = false; this.isEditingContest = false; this.editingContestId = null;
    this.contestEntityForm.reset();
  }

  onSubmitContestEntity(): void {
    if (!this.contestEntityForm.valid) return;
    const v = this.contestEntityForm.value;
    const startTime = new Date(v.start_time).toISOString();
    const endTime = new Date(v.end_time).toISOString();
    if (this.isEditingContest && this.editingContestId) {
      const contestIndex = this.contests.findIndex(c => c.id === this.editingContestId);
      let originalContest: Contest | null = null;
      if (contestIndex !== -1) {
        originalContest = { ...this.contests[contestIndex] };
        this.contests[contestIndex] = { ...this.contests[contestIndex], name: v.name, description: v.description || '', start_time: startTime, end_time: endTime };
        if (this.selectedContest?.id === this.editingContestId) { this.selectedContest = { ...this.contests[contestIndex] }; }
      }
      const editingId = this.editingContestId;
      this.cancelContestForm();
      this.contestAdminService.updateContest(editingId, v.name, v.description || '', startTime, endTime, false).subscribe({
        next: (updated) => {
          this.messageEmitted.emit({ msg: 'Contest updated', type: 'success' });
          const index = this.contests.findIndex(c => c.id === updated.id);
          if (index !== -1) { this.contests[index] = updated; } else { this.loadContests(); }
          if (this.selectedContest?.id === editingId) { this.selectedContest = updated; }
        },
        error: (err) => {
          if (contestIndex !== -1 && originalContest) { this.contests[contestIndex] = originalContest; if (this.selectedContest?.id === editingId) { this.selectedContest = { ...originalContest }; } }
          this.messageEmitted.emit({ msg: err.error?.error || 'Error updating contest', type: 'error' });
        }
      });
    } else {
      const tempContest: Contest = { id: 'temp-' + Date.now(), name: v.name, description: v.description || '', start_time: startTime, end_time: endTime, is_active: false };
      this.contests = [tempContest, ...this.contests];
      this.cancelContestForm();
      this.contestAdminService.createContest(v.name, v.description || '', startTime, endTime).subscribe({
        next: (created) => {
          this.messageEmitted.emit({ msg: 'Contest created', type: 'success' });
          const tempIndex = this.contests.findIndex(c => c.id === tempContest.id);
          if (tempIndex !== -1) { this.contests[tempIndex] = created; } else { this.loadContests(); }
        },
        error: (err) => {
          this.contests = this.contests.filter(c => c.id !== tempContest.id);
          this.messageEmitted.emit({ msg: err.error?.error || 'Error creating contest', type: 'error' });
        }
      });
    }
  }

  setActiveContest(contestId: string): void {
    this.contestAdminService.setActiveContest(contestId).subscribe({
      next: () => { this.messageEmitted.emit({ msg: 'Active contest updated', type: 'success' }); this.activeContestId = contestId; this.loadContestConfig(); },
      error: (err) => this.messageEmitted.emit({ msg: err.error?.error || 'Error setting active contest', type: 'error' })
    });
  }

  deleteContest(contest: Contest): void {
    this.confirmationModalService.show({ title: 'Delete Contest', message: `Are you sure you want to delete "${contest.name}"? This will also delete all rounds.`, confirmText: 'Delete', cancelText: 'Cancel' }).pipe(take(1)).subscribe(confirmed => {
      if (!confirmed) return;
      const wasSelected = this.selectedContest?.id === contest.id;
      const wasActive = this.activeContestId === contest.id;
      const originalContest = { ...contest };
      this.contests = this.contests.filter(c => c.id !== contest.id);
      if (wasSelected) { this.selectedContest = null; this.rounds = []; }
      if (wasActive) { this.activeContestId = null; }
      this.contestAdminService.deleteContest(contest.id).subscribe({
        next: () => { this.messageEmitted.emit({ msg: 'Contest deleted', type: 'success' }); this.loadContests(); if (wasActive) { this.loadContestConfig(); } },
        error: (err) => {
          this.contests = [...this.contests, originalContest];
          if (wasSelected) { this.selectContest(originalContest); }
          if (wasActive) { this.activeContestId = contest.id; this.loadContestConfig(); }
          this.loadContests();
          this.messageEmitted.emit({ msg: err.error?.error || 'Error deleting contest', type: 'error' });
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
    this.isCreatingRound = true; this.isEditingRound = false; this.editingRoundId = null;
    this.roundForm.reset({ name: '', description: '', order: this.rounds.length, visible_from: '', start_time: '', end_time: '' });
  }

  startEditRound(round: ContestRound): void {
    this.isCreatingRound = false; this.isEditingRound = true; this.editingRoundId = round.id;
    this.roundForm.patchValue({
      name: round.name, description: round.description || '', order: round.order,
      visible_from: round.visible_from ? new Date(round.visible_from).toISOString() : '',
      start_time: round.start_time ? new Date(round.start_time).toISOString() : '',
      end_time: round.end_time ? new Date(round.end_time).toISOString() : ''
    });
  }

  cancelRoundForm(): void {
    this.isCreatingRound = false; this.isEditingRound = false; this.editingRoundId = null;
    this.roundForm.reset();
  }

  onSubmitRound(): void {
    if (!this.roundForm.valid || !this.selectedContest) return;
    const v = this.roundForm.value;
    const visibleFrom = new Date(v.visible_from).toISOString();
    const startTime = new Date(v.start_time).toISOString();
    const endTime = new Date(v.end_time).toISOString();
    if (this.isEditingRound && this.editingRoundId) {
      const roundIndex = this.rounds.findIndex(r => r.id === this.editingRoundId);
      let originalRound: ContestRound | null = null;
      if (roundIndex !== -1) {
        originalRound = { ...this.rounds[roundIndex] };
        this.rounds[roundIndex] = { ...this.rounds[roundIndex], name: v.name, description: v.description || '', order: v.order, visible_from: visibleFrom, start_time: startTime, end_time: endTime };
        this.rounds.sort((a, b) => a.order - b.order);
        if (this.selectedRound?.id === this.editingRoundId) { const ni = this.rounds.findIndex(r => r.id === this.editingRoundId); if (ni !== -1) { this.selectedRound = { ...this.rounds[ni] }; } }
      }
      const editingId = this.editingRoundId;
      this.cancelRoundForm();
      this.contestAdminService.updateRound(this.selectedContest.id, editingId, v.name, v.description || '', v.order, visibleFrom, startTime, endTime).subscribe({
        next: (updated) => {
          this.messageEmitted.emit({ msg: 'Round updated', type: 'success' });
          const index = this.rounds.findIndex(r => r.id === updated.id);
          if (index !== -1) { this.rounds[index] = updated; this.rounds.sort((a, b) => a.order - b.order); } else { this.loadRounds(); }
        },
        error: (err) => {
          if (roundIndex !== -1 && originalRound) { this.rounds[roundIndex] = originalRound; this.rounds.sort((a, b) => a.order - b.order); if (this.selectedRound?.id === editingId) { this.selectedRound = { ...originalRound }; } }
          this.messageEmitted.emit({ msg: err.error?.error || 'Error updating round', type: 'error' });
        }
      });
    } else {
      const tempRound: ContestRound = { id: 'temp-' + Date.now(), contest_id: this.selectedContest.id, name: v.name, description: v.description || '', order: v.order, visible_from: visibleFrom, start_time: startTime, end_time: endTime };
      this.rounds = [...this.rounds, tempRound].sort((a, b) => a.order - b.order);
      this.cancelRoundForm();
      this.contestAdminService.createRound(this.selectedContest.id, v.name, v.description || '', v.order, visibleFrom, startTime, endTime).subscribe({
        next: (created) => {
          this.messageEmitted.emit({ msg: 'Round created', type: 'success' });
          const tempIndex = this.rounds.findIndex(r => r.id === tempRound.id);
          if (tempIndex !== -1) { this.rounds[tempIndex] = created; this.rounds.sort((a, b) => a.order - b.order); } else { this.loadRounds(); }
        },
        error: (err) => {
          this.rounds = this.rounds.filter(r => r.id !== tempRound.id);
          this.messageEmitted.emit({ msg: err.error?.error || 'Error creating round', type: 'error' });
        }
      });
    }
  }

  deleteRound(round: ContestRound): void {
    if (!this.selectedContest) return;
    this.confirmationModalService.show({ title: 'Delete Round', message: `Are you sure you want to delete "${round.name}"?`, confirmText: 'Delete', cancelText: 'Cancel' }).pipe(take(1)).subscribe(confirmed => {
      if (!confirmed) return;
      const wasSelected = this.selectedRound?.id === round.id;
      const originalRounds = [...this.rounds];
      this.rounds = this.rounds.filter(r => r.id !== round.id);
      if (wasSelected) { this.selectedRound = null; this.roundChallengeIds = []; this.availableChallengesForRound = []; }
      this.cancelRoundForm();
      this.contestAdminService.deleteRound(this.selectedContest!.id, round.id).subscribe({
        next: () => { this.messageEmitted.emit({ msg: 'Round deleted', type: 'success' }); this.loadRounds(); },
        error: (err) => {
          this.rounds = originalRounds;
          if (wasSelected) { this.selectedRound = round; this.loadRoundChallenges(); }
          this.messageEmitted.emit({ msg: err.error?.error || 'Error deleting round', type: 'error' });
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

  formatDateForDisplay(dateStr: string): string {
    if (!dateStr) return '';
    return new Date(dateStr).toLocaleString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' });
  }

  // Round-Challenge attachment
  selectRound(round: ContestRound): void {
    this.selectedRound = this.selectedRound?.id === round.id ? null : round;
    if (this.selectedRound) {
      this.selectedAttachedChallenges.clear(); this.selectedAvailableChallenges.clear();
      this.loadRoundChallenges();
    } else {
      this.roundChallengeIds = []; this.availableChallengesForRound = [];
      this.selectedAttachedChallenges.clear(); this.selectedAvailableChallenges.clear();
    }
  }

  loadRoundChallenges(): void {
    if (!this.selectedContest || !this.selectedRound) return;
    this.isLoadingRoundChallenges = true;
    this.contestAdminService.getRoundChallenges(this.selectedContest.id, this.selectedRound.id).subscribe({
      next: (ids) => {
        this.roundChallengeIds = ids || []; this.isLoadingRoundChallenges = false;
        this.computeAvailableChallenges(); this.selectedAttachedChallenges.clear(); this.selectedAvailableChallenges.clear();
      },
      error: () => { this.roundChallengeIds = []; this.isLoadingRoundChallenges = false; }
    });
  }

  computeAvailableChallenges(): void {
    const attachedSet = new Set(this.roundChallengeIds);
    this.availableChallengesForRound = this.challenges.filter(ch => !attachedSet.has(ch.id));
  }

  attachChallenge(challengeId: string): void { this.attachChallenges([challengeId]); }

  attachChallenges(challengeIds: string[]): void {
    if (!this.selectedContest || !this.selectedRound || challengeIds.length === 0) return;
    const newIds = challengeIds.filter(id => !this.roundChallengeIds.includes(id));
    this.roundChallengeIds = [...this.roundChallengeIds, ...newIds];
    this.computeAvailableChallenges(); this.selectedAvailableChallenges.clear();
    this.contestAdminService.attachChallenges(this.selectedContest.id, this.selectedRound.id, challengeIds).subscribe({
      next: () => { this.messageEmitted.emit({ msg: challengeIds.length === 1 ? 'Challenge attached to round' : `${challengeIds.length} challenges attached to round`, type: 'success' }); this.loadRoundChallenges(); },
      error: (err) => { this.roundChallengeIds = this.roundChallengeIds.filter(id => !newIds.includes(id)); this.computeAvailableChallenges(); this.messageEmitted.emit({ msg: err.error?.error || 'Error attaching challenges', type: 'error' }); }
    });
  }

  detachChallenge(challengeId: string): void { this.detachChallenges([challengeId]); }

  detachChallenges(challengeIds: string[]): void {
    if (!this.selectedContest || !this.selectedRound || challengeIds.length === 0) return;
    this.confirmationModalService.show({ title: challengeIds.length === 1 ? 'Detach Challenge' : 'Detach Challenges', message: challengeIds.length === 1 ? 'Remove this challenge from the round?' : `Remove ${challengeIds.length} challenges from the round?`, confirmText: 'Remove', cancelText: 'Cancel' }).pipe(take(1)).subscribe(confirmed => {
      if (!confirmed) return;
      const hadChallenges = challengeIds.filter(id => this.roundChallengeIds.includes(id));
      this.roundChallengeIds = this.roundChallengeIds.filter(id => !challengeIds.includes(id));
      this.computeAvailableChallenges(); this.selectedAttachedChallenges.clear();
      this.contestAdminService.detachChallenges(this.selectedContest!.id, this.selectedRound!.id, challengeIds).subscribe({
        next: () => { this.messageEmitted.emit({ msg: challengeIds.length === 1 ? 'Challenge detached from round' : `${challengeIds.length} challenges detached from round`, type: 'success' }); this.loadRoundChallenges(); },
        error: (err) => { this.roundChallengeIds = [...this.roundChallengeIds, ...hadChallenges]; this.computeAvailableChallenges(); this.messageEmitted.emit({ msg: err.error?.error || 'Error detaching challenges', type: 'error' }); }
      });
    });
  }

  toggleAttachedChallenge(id: string): void { if (this.selectedAttachedChallenges.has(id)) { this.selectedAttachedChallenges.delete(id); } else { this.selectedAttachedChallenges.add(id); } }
  toggleAvailableChallenge(id: string): void { if (this.selectedAvailableChallenges.has(id)) { this.selectedAvailableChallenges.delete(id); } else { this.selectedAvailableChallenges.add(id); } }
  isAttachedChallengeSelected(id: string): boolean { return this.selectedAttachedChallenges.has(id); }
  isAvailableChallengeSelected(id: string): boolean { return this.selectedAvailableChallenges.has(id); }
  selectAllAttached(): void { this.roundChallengeIds.forEach(id => this.selectedAttachedChallenges.add(id)); }
  deselectAllAttached(): void { this.selectedAttachedChallenges.clear(); }
  selectAllAvailable(): void { this.availableChallengesForRound.forEach(ch => this.selectedAvailableChallenges.add(ch.id)); }
  deselectAllAvailable(): void { this.selectedAvailableChallenges.clear(); }
  areAllAttachedSelected(): boolean { return this.roundChallengeIds.length > 0 && this.roundChallengeIds.every(id => this.selectedAttachedChallenges.has(id)); }
  areAllAvailableSelected(): boolean { return this.availableChallengesForRound.length > 0 && this.availableChallengesForRound.every(ch => this.selectedAvailableChallenges.has(ch.id)); }
  bulkAttachSelected(): void { const ids = Array.from(this.selectedAvailableChallenges); if (ids.length > 0) { this.attachChallenges(ids); } }
  bulkDetachSelected(): void { const ids = Array.from(this.selectedAttachedChallenges); if (ids.length > 0) { this.detachChallenges(ids); } }

  getChallengeTitle(id: string): string { const ch = this.challenges.find(c => c.id === id); return ch ? ch.title : id; }
  getChallengeCategory(id: string): string { const ch = this.challenges.find(c => c.id === id); return ch ? this.getCategoryLabel(ch.category) : ''; }
  getChallengeDifficulty(id: string): string { const ch = this.challenges.find(c => c.id === id); return ch ? ch.difficulty : ''; }
  getChallengePoints(id: string): number { const ch = this.challenges.find(c => c.id === id); return ch ? ch.current_points : 0; }
  getCategoryLabel(value: string): string { const cat = this.categories.find(c => c.value === value); return cat ? cat.label : value; }
}
