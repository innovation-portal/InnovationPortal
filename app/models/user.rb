class User < ApplicationRecord
  has_secure_password

  validates :email, presence: true, uniqueness: true

  def self.find_or_create_by_omniauth(auth_hash)
    # query user table to find email that matches GitHub email
    # return instance of user found or instance of user just created
    # if user created, pass the newly instantiated user to user block, set the user's password, save the user and then return the new user

    self.where(email: auth_hash["info"]["email"]).first_or_create do |user|
      # creates random password
      user.password = SecureRandom.hex
    end
  end
end
