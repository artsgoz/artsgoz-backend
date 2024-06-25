import mongoose, { Schema } from 'mongoose';

const staffSchema = mongoose.Schema({
    abbr: {
        type: String,
        required: true,        
    },
    name: {
        type: String,
        required: true,
    },
    dept: {
        type: String,
        required: true,
    },
},
{ 
    timestamps: {
        createdAt: 'publishedAt'
    } 

})

export default mongoose.model("staffs", staffSchema);