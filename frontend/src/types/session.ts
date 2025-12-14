export type SessionDTO = {
  id: number;
  date: string;
  start_time: number;
  end_time: number;
  time_range: string;
  primary_app: string;
  category: string;
  summary: string;
  skills_involved: string[];
  diff_count: number;
  browser_count: number;
};

