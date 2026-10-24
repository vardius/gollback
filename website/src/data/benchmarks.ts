// Results behind the benchmark charts and tables. Rerun with the commands on the Benchmarks page and update together.

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
		{ scenario: '1 function', before: { ns: 795.2, allocs: 13 }, after: { ns: 348.6, allocs: 4 } },
		{ scenario: '4 functions', before: { ns: 2046, allocs: 22 }, after: { ns: 1425, allocs: 7 } },
		{ scenario: '16 functions', before: { ns: 6574, allocs: 58 }, after: { ns: 4446, allocs: 19 } },
		{ scenario: '4 functions, all fail', before: { ns: 2336, allocs: 22 }, after: { ns: 1167, allocs: 11 } },
		{ scenario: '4 functions, 16 callers at once', before: { ns: 434.9, allocs: 22 }, after: { ns: 295.6, allocs: 7 } },
	],
	all: [
		{ scenario: '1 function', before: { ns: 293.4, allocs: 5 }, after: { ns: 17.7, allocs: 2 } },
		{ scenario: '4 functions', before: { ns: 1040, allocs: 11 }, after: { ns: 759.2, allocs: 6 } },
		{ scenario: '16 functions', before: { ns: 3411, allocs: 35 }, after: { ns: 2826, allocs: 18 } },
		{ scenario: '4 functions, 16 callers at once', before: { ns: 316.3, allocs: 11 }, after: { ns: 189.9, allocs: 6 } },
	],
	retry: [{ scenario: '5 failed attempts', before: { ns: 25.4, allocs: 0 }, after: { ns: 14.1, allocs: 0 } }],
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
	{ benchmark: 'Race/1', what: 'Race one function', ns: 338.6, bytes: 352, allocs: 4 },
	{ benchmark: 'Race/4', what: 'One winner, three losers that wait to be cancelled', ns: 1463, bytes: 445, allocs: 7 },
	{ benchmark: 'Race/16', what: 'One winner, fifteen losers', ns: 4871, bytes: 814, allocs: 19 },
	{ benchmark: 'RaceAllFail', what: 'Four functions that all fail, errors joined', ns: 1157, bytes: 600, allocs: 11 },
	{ benchmark: 'RaceParallel', what: 'Race/4 from 16 goroutines at once', ns: 311.8, bytes: 447, allocs: 7 },
	{ benchmark: 'All/1', what: 'One function, run on the caller\'s goroutine', ns: 15.8, bytes: 24, allocs: 2 },
	{ benchmark: 'All/4', what: 'Four functions', ns: 769.7, bytes: 304, allocs: 6 },
	{ benchmark: 'All/16', what: 'Sixteen functions', ns: 2873, bytes: 976, allocs: 18 },
	{ benchmark: 'AllParallel', what: 'All/4 from 16 goroutines at once', ns: 203.4, bytes: 304, allocs: 6 },
	{ benchmark: 'Retry/1', what: 'Succeeds on the first attempt', ns: 2.49, bytes: 0, allocs: 0 },
	{ benchmark: 'Retry/5', what: 'Succeeds on the fifth attempt', ns: 10.98, bytes: 0, allocs: 0 },
];
