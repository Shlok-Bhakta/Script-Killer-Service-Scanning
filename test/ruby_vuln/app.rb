require 'sinatra'
require 'nokogiri'

get '/' do
  "Hello World"
end

post '/parse' do
  doc = Nokogiri::XML(request.body.read)
  doc.to_s
end
