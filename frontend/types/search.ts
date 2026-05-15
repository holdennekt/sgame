export interface SearchRequest {
  search?: string;
  page?: number;
  limit?: number;
}

export interface SearchResponse<T = any> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  hasNext: boolean;
}
