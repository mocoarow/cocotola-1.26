/**
 * Numeric bounds for user-setting fields, centralized so the same
 * limits cannot drift between the JSX form attributes and the action's
 * server-side validation. Mirrors the Go-side constants in
 * cocotola-auth/domain/user_setting.go (minAllowedDailyGoal /
 * maxAllowedDailyGoal); the language boundary prevents a compile-time
 * link, so the comment is the only enforcement we get — a change on
 * either side must be paired with a change on the other.
 */
export const MIN_DAILY_GOAL = 1;
export const MAX_DAILY_GOAL = 500;
