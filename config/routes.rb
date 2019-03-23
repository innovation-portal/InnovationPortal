Rails.application.routes.draw do

  root 'welcome#home'

  get 'sessions/new'
  get 'sessions/create'
  get 'sessions/destroy'
  resources :users

  get '/signup' => 'users#new'
  post '/signup' => 'users#create'

  get '/login' => 'sessions#new'
  post '/login' => 'sessions#create'
  get '/auth/:github/callback' => 'sessions#create'

  delete '/logout' => 'sessions#destroy'
  # For details on the DSL available within this file, see http://guides.rubyonrails.org/routing.html
end
