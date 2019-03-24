class ProjectsController < ApplicationController

  def index
    @projects = Project.all

    respond_to do |f|
      f.html {render :index}
      f.json {render json: @projects}
    end
  end

  # def new
  #   @project = Project.new
  # end

  # def create
  #   @project = Project.new

  #   if @project.save
  #     render json: @project, status: 201
  #   end
  # end

end



