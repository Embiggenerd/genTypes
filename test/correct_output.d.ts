export interface ErrorData {
  status_code?: number /* int */;
  message?: string;
  public: boolean;
  CreatedAt: Date;
  Deleted: boolean;
}

export type Errors = ErrorData[];

