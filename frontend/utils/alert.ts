import Swal from 'sweetalert2';

export const showSuccessAlert = (message: string) => {
  Swal.fire({
    icon: 'success',
    title: message,
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
  });
};

export const showErrorAlert = (message: string) => {
  Swal.fire({
    icon: 'error',
    title: message,
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
  });
};

export const showInfoAlert = (message: string) => {
  Swal.fire({
    icon: 'info',
    title: message,
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
  });
};

export const showWarningAlert = (message: string) => {
  Swal.fire({
    icon: 'warning',
    title: message,
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
  });
};

export const showConfirmDialog = async (title: string, text: string): Promise<boolean> => {
  const result = await Swal.fire({
    title,
    text,
    icon: 'warning',
    showCancelButton: true,
    confirmButtonColor: '#3085d6',
    cancelButtonColor: '#d33',
    confirmButtonText: 'Yes',
    cancelButtonText: 'No',
  });
  return result.isConfirmed;
}; 