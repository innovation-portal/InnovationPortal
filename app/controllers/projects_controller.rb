class ProjectsController < ApplicationController

    def index
      @projects = Project.all

      respond_to do |f|
        f.html {render :index}
        f.json {render json: @projects}
      end
    end

end



