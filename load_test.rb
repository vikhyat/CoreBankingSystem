require 'open-uri'

HAPROXY = "10.6.1.160"
PORT    = 9292

def random_account
  rand 10**7
end

def random_amount
  rand 10**4
end

def random_request_url
  [
    "http://#{HAPROXY}:#{PORT}/deposit?account=#{random_account}&amount=#{random_amount}",
    "http://#{HAPROXY}:#{PORT}/withdraw?account=#{random_account}&amount=#{random_amount}",
    "http://#{HAPROXY}:#{PORT}/transfer?source=#{random_account}&destination=#{random_account}&amount=#{random_amount}"
  ].sample
end

ts = []
4.times do
  ts.push(Thread.new {
    loop do
      url = random_request_url
      open(url).read rescue nil
      puts url
    end
  })
end
ts.each {|t| t.join }
