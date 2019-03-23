class CreateProjects < ActiveRecord::Migration[5.2]
  def change
    create_table :projects do |t|
      t.string "name"
      t.string "tags", array:true
      t.string "tag_line"
      t.string "members", array:true
      t.string "photo"
      t.string "application_area", array:true
      t.boolean :winner
      t.string "winner_type"
      t.string "hackathon"
      t.integer :year
      t.timestamps
    end
  end
end
