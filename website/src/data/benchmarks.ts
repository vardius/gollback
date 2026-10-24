// Results behind the benchmark charts and tables, updated together. The suite reruns with the commands on the
// Benchmarks page; the comparison comes from a harness that runs the old and new versions side by side.

export const environment = 'Apple M4 Max, Go 1.27.0, median of 5 runs';

/** The two versions the charts compare. */
export const versions = { before: 'v1', after: 'v2' } as const;

export type Group = 'race' | 'all' | 'retry';

export interface Result {
	ns: number;
	allocs: number;
}

export interface Comparison {
	scenario: string;
	before: Result;
	after: Result;
}

/**
 * v1 against v2 on the same harness, one call per op. The functions return at once, so what is left is
 * gollback's own cost. A race has one winner; the other functions wait to be cancelled.
 */
export const comparison: Record<Group, Comparison[]> = {
	race: [
		{ scenario: '1 function', before: { ns: 762.3, allocs: 13 }, after: { ns: 347.5, allocs: 5 } },
		{ scenario: '4 functions', before: { ns: 1950, allocs: 22 }, after: { ns: 1350, allocs: 8 } },
		{ scenario: '16 functions', before: { ns: 5630, allocs: 58 }, after: { ns: 3940, allocs: 20 } },
		{ scenario: '4 functions, all fail', before: { ns: 2190, allocs: 22 }, after: { ns: 1080, allocs: 11 } },
		{ scenario: '4 functions, 16 callers at once', before: { ns: 445.6, allocs: 22 }, after: { ns: 284.4, allocs: 8 } },
	],
	all: [
		{ scenario: '1 function', before: { ns: 284.5, allocs: 5 }, after: { ns: 15.2, allocs: 2 } },
		{ scenario: '4 functions', before: { ns: 950.1, allocs: 11 }, after: { ns: 718.9, allocs: 6 } },
		{ scenario: '16 functions', before: { ns: 3080, allocs: 35 }, after: { ns: 2660, allocs: 18 } },
		{ scenario: '4 functions, 16 callers at once', before: { ns: 313.2, allocs: 11 }, after: { ns: 191.4, allocs: 6 } },
	],
	retry: [{ scenario: '5 failed attempts', before: { ns: 24.6, allocs: 0 }, after: { ns: 13.4, allocs: 0 } }],
};

export interface SuiteResult {
	benchmark: string;
	what: string;
	ns: number;
	bytes: number;
	allocs: number;
}

/** v2's own benchmark suite, `go test -bench .`. */
export const suite: SuiteResult[] = [
	{ benchmark: 'Race/1', what: 'Race one function', ns: 343.7, bytes: 352, allocs: 5 },
	{ benchmark: 'Race/4', what: 'One winner, three losers that wait to be cancelled', ns: 1370, bytes: 448, allocs: 8 },
	{ benchmark: 'Race/16', what: 'One winner, fifteen losers', ns: 4050, bytes: 834, allocs: 20 },
	{ benchmark: 'RaceAllFail', what: 'Four functions that all fail, errors joined', ns: 1100, bytes: 600, allocs: 11 },
	{ benchmark: 'RaceParallel', what: 'Race/4 from 16 goroutines at once', ns: 273.1, bytes: 448, allocs: 8 },
	{ benchmark: 'All/1', what: 'One function, run on the caller\'s goroutine', ns: 13.7, bytes: 24, allocs: 2 },
	{ benchmark: 'All/4', what: 'Four functions', ns: 741, bytes: 304, allocs: 6 },
	{ benchmark: 'All/16', what: 'Sixteen functions', ns: 2760, bytes: 976, allocs: 18 },
	{ benchmark: 'AllParallel', what: 'All/4 from 16 goroutines at once', ns: 181.4, bytes: 304, allocs: 6 },
	{ benchmark: 'Retry/1', what: 'Succeeds on the first attempt', ns: 2.5, bytes: 0, allocs: 0 },
	{ benchmark: 'Retry/5', what: 'Succeeds on the fifth attempt', ns: 10.7, bytes: 0, allocs: 0 },
];
