output "vpc_id" {
  value = google_compute_network.vpc.id
}

output "gke_subnet_id" {
  value = google_compute_subnetwork.gke_subnet.id
}