export interface ErrorResponse {
  error: string;
}

export class HttpError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}
