import React, { useState, ChangeEvent, FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import authService from "../../services/auth";
import { RegisterCredentials } from "../../types/auth.types";

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const Register: React.FC = () => {
	const navigate = useNavigate();
	const [formData, setFormData] = useState<RegisterCredentials>({
		email: "",
		password: "",
		name: "",
	});
	const [emailError, setEmailError] = useState<string>("");
	const [passwordError, setPasswordError] = useState<string>("");
	const [error, setError] = useState<string>("");
	const [isLoading, setIsLoading] = useState<boolean>(false);

	const handleChange = (e: ChangeEvent<HTMLInputElement>): void => {
		const { name, value } = e.target;

		setFormData(prev => ({
			...prev,
			[name]: value,
		}));

		if (name === "email") {
			setEmailError(!emailRegex.test(value) ? "Please enter a valid email address." : "");
		}

		if (name === "password") {
			setPasswordError(value.length < 8 ? "Password must be at least 8 characters long." : "");
		}
	};

	const handleSubmit = async (e: FormEvent<HTMLFormElement>): Promise<void> => {
		e.preventDefault();
		setError("");

		// Final client-side checks
		if (!emailRegex.test(formData.email)) {
			setEmailError("Please enter a valid email address.");
			return;
		}
		if (formData.password.length < 8) {
			setPasswordError("Password must be at least 8 characters long.");
			return;
		}

		setIsLoading(true);
		try {
			const response = await authService.register(formData);
			if (!localStorage.getItem("userEmail")) {
				localStorage.setItem("userEmail", response.email);
			}
			navigate("/login");
		} catch (err) {
			setError(err instanceof Error ? err.message : "Registration failed. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	const hasFormError = !!emailError || !!passwordError;

	return (
		<div className="min-h-screen flex items-center justify-center bg-gray-50">
			<div className="max-w-md w-full space-y-8 p-8 bg-white rounded-lg shadow">
				<h2 className="text-center text-3xl font-bold">Create Account</h2>

				{error && (
					<div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded" role="alert">
						{error}
					</div>
				)}

				<form onSubmit={handleSubmit} className="mt-8 space-y-6">
					<div className="space-y-4">
						<div>
							<label htmlFor="name" className="block text-sm font-medium">
								Full Name
							</label>
							<input
								id="name"
								type="text"
								name="name"
								value={formData.name}
								onChange={handleChange}
								required
								disabled={isLoading}
								className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
							/>
						</div>

						<div>
							<label htmlFor="email" className="block text-sm font-medium">
								Email
							</label>
							<input
								id="email"
								type="email"
								name="email"
								value={formData.email}
								onChange={handleChange}
								required
								disabled={isLoading}
								className={`mt-1 block w-full px-3 py-2 border rounded-md ${
									emailError ? "border-red-500" : "border-gray-300"
								}`}
							/>
							{emailError && <p className="text-red-500 text-xs mt-1">{emailError}</p>}
						</div>

						<div>
							<label htmlFor="password" className="block text-sm font-medium">
								Password
							</label>
							<input
								id="password"
								type="password"
								name="password"
								value={formData.password}
								onChange={handleChange}
								required
								disabled={isLoading}
								className={`mt-1 block w-full px-3 py-2 border rounded-md ${
									passwordError ? "border-red-500" : "border-gray-300"
								}`}
							/>
							{passwordError && <p className="text-red-500 text-xs mt-1">{passwordError}</p>}
						</div>
					</div>

					<button
						type="submit"
						disabled={isLoading || hasFormError}
						className={`w-full py-2 px-4 bg-blue-600 text-white rounded-md hover:bg-blue-700 ${
							isLoading || hasFormError ? "opacity-50 cursor-not-allowed" : ""
						}`}>
						{isLoading ? (
							<span className="flex items-center justify-center">
								<svg
									className="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24">
									<circle
										className="opacity-25"
										cx="12"
										cy="12"
										r="10"
										stroke="currentColor"
										strokeWidth="4"
									/>
									<path
										className="opacity-75"
										fill="currentColor"
										d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
									/>
								</svg>
								Creating Account...
							</span>
						) : (
							"Register"
						)}
					</button>
				</form>

				<div className="text-center mt-4">
					<p className="text-sm">
						Already have an account?{" "}
						<button onClick={() => navigate("/login")} className="text-blue-600 hover:underline">
							Sign in
						</button>
					</p>
				</div>
			</div>
		</div>
	);
};

export default Register;
