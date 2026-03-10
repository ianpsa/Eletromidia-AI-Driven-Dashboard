export interface ObjectItem {
  id: string;
  size: number;
  content_type?: string;
  updated_at: string;
}

export interface DisplayItem {
  type: "folder" | "file";
  id: string;
  name: string;
  size: number;
  content_type?: string;
  updated_at: string;
}

export interface BreadcrumbItem {
  label: string;
  path: string;
}
