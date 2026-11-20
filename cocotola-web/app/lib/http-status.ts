const STATUS_SERVER_ERROR_MIN = 500;

export function isServerErrorStatus(status: number): boolean {
  return status >= STATUS_SERVER_ERROR_MIN;
}
