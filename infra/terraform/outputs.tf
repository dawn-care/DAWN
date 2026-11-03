output "public_ip" {
  description = "Static public IPv4 address for this facility."
  value       = aws_eip.facility.public_ip
}

output "instance_id" {
  value = aws_instance.facility.id
}

output "storage_volume_id" {
  description = "Encrypted disk; retained if the instance is terminated."
  value       = aws_instance.facility.root_block_device[0].volume_id
}

output "ssh_command" {
  value = "ssh ubuntu@${aws_eip.facility.public_ip}"
}
