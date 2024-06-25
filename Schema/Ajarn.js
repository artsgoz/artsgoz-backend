import mongoose, { Schema } from "mongoose";

const ajarnSchema = mongoose.Schema({

    personal_info: {
        fullname: {
            type: String,
            required: true,
        },
        username: {
            type: String,
            unique: false,
        },
        dept: {
            type: String,
            unique: false,
        },
        bio: {
            type: String,
            default: "",
        },        
        profile_img: {
            type: String,
            default: () => {
                return `https://www.arts.chula.ac.th/goz/asset/user_image.png`
            } 
        },
    },
    social_links: {
        youtube: {
            type: String,
            default: "",
        },
        instagram: {
            type: String,
            default: "",
        },
        facebook: {
            type: String,
            default: "",
        },
        twitter: {
            type: String,
            default: "",
        },
        github: {
            type: String,
            default: "",
        },
        website: {
            type: String,
            default: "",
        }
    },
    account_info:{
        total_posts: {
            type: Number,
            default: 0
        },
        total_reads: {
            type: Number,
            default: 0
        },
    },
    google_auth: {
        type: Boolean,
        default: false
    },
    blogs: {
        type: [ Schema.Types.ObjectId ],
        ref: 'blogs',
        default: [],
    }

}, 
{ 
    timestamps: {
        createdAt: 'joinedAt'
    } 

})

export default mongoose.model("ajarns", ajarnSchema);