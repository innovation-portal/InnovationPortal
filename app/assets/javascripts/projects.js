$(function() {
  getProjects();
  $("#example").DataTable();
});

const getProjects = () => {
  $.ajax({
    type: "get",
    url: "https://api.hackportal.net/v1/api/projects",
    dataType: "json",
    success: function(response) {
      displayProjects(response);
    }
  });
};

const displayProjects = response => {
  response.forEach(project => {
    const newProject = new Project(project);
    const newProjectHtml = newProject.projectHTML();

    document.getElementById("ajax-projects").innerHTML += newProjectHtml;
  });
};

class Project {
  constructor(obj) {
    this.ID = obj.ID;
    this.Name = obj.Name;
    this.Description = obj.Description;
    this.Tags = obj.Tags;
    this.Members = obj.Members;
    this.Photo = obj.Photo;
    this.ApplicationArea = obj.ApplicationArea;
    this.Winner = obj.Winner;
    this.WinnerType = obj.WinnerType;
    this.Hackathon = obj.Hackathon;
    this.Year = obj.Year;
  }
}

Project.prototype.projectHTML = function() {
  return `
    <tr>
    <td>${this.Name}</td>
    <td>${this.Description}</td>
    <td>${this.Tags}</td>
    <td>${this.Members}</td>
    <td>${this.Photo}</td>
    <td>${this.Winner}</td>
    <td>${this.WinnerType}</td>
    <td>${this.Hackathon}</td>
    <td>${this.Year}</td>
    </tr>
  `;
};
