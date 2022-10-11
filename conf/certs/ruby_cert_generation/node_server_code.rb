#Node server is postanalytics exporter

require 'openssl'

# Node generate
node_server_root = OpenSSL::PKey::RSA.new 4096

open 'node_server_private_key.pem', 'w', 0400 do |io|
  io.write node_server_root #.export(cipher, pass_phrase)
end

# create CSR

unsigned_node_instance_cert = OpenSSL::PKey.read File.read 'node_server_private_key.pem'


ca_name = OpenSSL::X509::Name.parse '/CN=server/DC=server-example' # replace with wallarm things
csr = OpenSSL::X509::Request.new
csr.version = 0
csr.subject = ca_name
csr.public_key = node_server_root.public_key
csr.sign node_server_root, OpenSSL::Digest::SHA256.new

open 'csr_server.pem', 'w' do |io|
  io.write csr.to_pem
end

