# AI Avatar Generator

A full-stack application that allows users to upload personal images and generate stylized AI avatars using Hugging Face's Stable Diffusion models. This project utilizes a Go backend for handling business logic and concurrency, React for the user interface, and Supabase for authentication and data storage.

## Features

- **User Authentication:** Secure passwordless login using Supabase Auth (Magic Links).
- **Image Upload:** Direct upload to Supabase Storage with public URL generation.
- **AI Generation:** Integration with Hugging Face Inference API for image-to-image style transfer.
- **Asynchronous Processing:** Go routines handle long-running AI tasks without blocking the user interface.



