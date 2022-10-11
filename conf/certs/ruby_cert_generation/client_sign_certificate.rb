# client sign certificate

require 'openssl'

# sign with wallarm api root
node_api_root_key = OpenSSL::PKey.read File.read 'node_root_private_key.pem'
node_api_root_cert = OpenSSL::X509::Certificate.new File.read 'node_root_public_key.pem'

csr = OpenSSL::X509::Request.new File.read 'csr_client.pem'

raise 'CSR can not be verified' unless csr.verify csr.public_key

csr_cert = OpenSSL::X509::Certificate.new
csr_cert.serial = 0
csr_cert.version = 2
csr_cert.not_before = Time.now
csr_cert.not_after = Time.now + 60*60*24*365*10

csr_cert.subject = csr.subject
csr_cert.public_key = csr.public_key
csr_cert.issuer = node_api_root_cert.subject

extension_factory = OpenSSL::X509::ExtensionFactory.new
extension_factory.subject_certificate = csr_cert
extension_factory.issuer_certificate = node_api_root_cert

csr_cert.add_extension extension_factory.create_extension('basicConstraints', 'CA:FALSE')
csr_cert.add_extension extension_factory.create_extension('keyUsage', 'keyEncipherment,dataEncipherment,digitalSignature')
csr_cert.add_extension extension_factory.create_extension('subjectKeyIdentifier', 'hash')
# csr_cert.add_extension extension_factory.create_extension('extKeyUsage', 'clientAuth')

csr_cert.sign node_api_root_key, OpenSSL::Digest.new('SHA1')

open 'node_client_cert.pem', 'w' do |io|
  io.write csr_cert.to_pem
end
