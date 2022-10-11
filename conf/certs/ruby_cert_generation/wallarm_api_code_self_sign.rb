require 'openssl'



# API generate
node_api_root = OpenSSL::PKey::RSA.new 4096
# pass_phrase = 'wallarm rules'

# cipher = OpenSSL::Cipher.new 'aes-256-cbc'

open 'node_root_private_key.pem', 'w', 0400 do |io|
  io.write node_api_root #.export(cipher, pass_phrase)
end

# API self sign

ca_name = OpenSSL::X509::Name.parse '/CN=ca/DC=example' # replace with wallarm?

ca_cert = OpenSSL::X509::Certificate.new
ca_cert.serial = 0
ca_cert.version = 2
ca_cert.not_before = Time.now
ca_cert.not_after = Time.now + 60*60*24*365*10

ca_cert.public_key = node_api_root.public_key
ca_cert.subject = ca_name
ca_cert.issuer = ca_name

extension_factory = OpenSSL::X509::ExtensionFactory.new
extension_factory.subject_certificate = ca_cert
extension_factory.issuer_certificate = ca_cert

ca_cert.add_extension extension_factory.create_extension('subjectKeyIdentifier', 'hash')
ca_cert.add_extension extension_factory.create_extension('basicConstraints', 'CA:TRUE', true)
ca_cert.add_extension extension_factory.create_extension('keyUsage', 'cRLSign,keyCertSign', true)

ca_cert.sign node_api_root, OpenSSL::Digest::SHA256.new

open 'node_root_public_key.pem', 'w', 0400 do |io|
  io.write ca_cert.to_pem
end
