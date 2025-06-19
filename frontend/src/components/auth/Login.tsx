import React, { useState, ChangeEvent, FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import authService from "../../services/auth";
import { LoginCredentials } from "../../types/auth.types";
import { ROUTES } from "../../constants";

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const Login: React.FC = () => {
	const navigate = useNavigate();
	const [formData, setFormData] = useState<LoginCredentials>({
		email: "",
		password: "",
	});
	const [emailError, setEmailError] = useState<string>("");
	const [passwordError, setPasswordError] = useState<string>("");
	const [error, setError] = useState<string>("");
	const [isLoading, setIsLoading] = useState<boolean>(false);

	const handleChange = (e: ChangeEvent<HTMLInputElement>): void => {
		const { name, value } = e.target;
		setFormData(prev => ({ ...prev, [name]: value }));

		if (name === "email") {
			setEmailError(!emailRegex.test(value) ? "Enter a valid email address." : "");
		}
		if (name === "password") {
			setPasswordError(value.length < 8 ? "Password must be at least 8 characters." : "");
		}
	};

	const handleSubmit = async (e: FormEvent<HTMLFormElement>): Promise<void> => {
		e.preventDefault();
		setError("");

		// Final client-side checks
		if (!emailRegex.test(formData.email)) {
			setEmailError("Enter a valid email address.");
			return;
		}
		if (formData.password.length < 8) {
			setPasswordError("Password must be at least 8 characters.");
			return;
		}

		setIsLoading(true);
		try {
			const response = await authService.login(formData);
			if (response) {
				if (!localStorage.getItem("userEmail")) {
					localStorage.setItem("userEmail", response.email);
				}
				navigate(ROUTES.HOME, { replace: true });
			}
		} catch (err) {
			// existing refresh-token logic...
			if (err instanceof Error && err.message === "Token expired") {
				try {
					const refreshResponse = await authService.refreshToken();
					if (refreshResponse) {
						await authService.login(formData);
						navigate(ROUTES.HOME, { replace: true });
					} else {
						setError("Session expired. Please log in again.");
					}
				} catch {
					setError("Login failed. Please check your credentials.");
				}
			} else {
				setError(err instanceof Error ? err.message : "Login failed. Please check your credentials.");
			}
		} finally {
			setIsLoading(false);
		}
	};

	const hasValidationError = !!emailError || !!passwordError;

	return (
		<div className="min-h-screen flex items-center justify-center bg-gray-50">
			<div className="max-w-md w-full space-y-8 p-8 bg-white rounded-lg shadow">
				<h2 className="text-center text-3xl font-bold">Sign In</h2>

				{error && (
					<div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded" role="alert">
						{error}
					</div>
				)}

				<form onSubmit={handleSubmit} className="mt-8 space-y-6">
					<div className="space-y-4">
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
						disabled={isLoading || hasValidationError}
						className={`w-full py-2 px-4 bg-blue-600 text-white rounded-md hover:bg-blue-700 ${
							isLoading || hasValidationError ? "opacity-50 cursor-not-allowed" : ""
						}`}>
						{isLoading ? "Signing in..." : "Sign In"}
					</button>

					<p className="text-center mt-4 text-sm">
						Don’t have an account?{" "}
						<Link to="/register" className="text-blue-600 hover:underline">
							Register here
						</Link>
					</p>
				</form>
			</div>
		</div>
	);
};

export default Login;
